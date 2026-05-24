package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	spinnerInterval = 500 * time.Millisecond
)

type spinnerTickMsg struct{}
type dataTickMsg struct{}

// ViewMode is the primary TUI screen when not in describe view.
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewPicker
	ViewFilter
)

// DeletePhase is the delete confirmation step in the main list view.
type DeletePhase int

const (
	DeletePhaseNone DeletePhase = iota
	DeletePhaseConfirm
)

// PortForwardToggler starts port-forwards when any are down, or stops all when all are up.
type PortForwardToggler func() (info string, err error)

// Options configures the TUI session.
type Options struct {
	Fetcher            EntryFetcher
	Patcher            Patcher
	Deleter            Deleter
	ContextProvider    ContextProvider
	PortForwardToggler PortForwardToggler
	RefreshSecs        int
	NoColor            bool
}

// Model is the bubbletea model for the HyperFleet TUI.
type Model struct {
	opts Options

	width  int
	height int

	snapshot Snapshot
	selected int
	tick     int
	secsLeft int
	nextRefresh time.Time
	context  ContextInfo

	viewMode ViewMode

	detailOpen     bool
	detailScroll   int
	detailFormat   DetailFormat
	detailSerialized DetailFormat

	pickerKind     FilterKind
	pickerItems    []string
	pickerSelected int

	filterKind    FilterKind
	filterKey     string
	filterHeaders []string
	filterRows    [][]string
	filterScroll  int

	tableOffsetY int
	patchMode      bool
	patching       bool
	deletePhase    DeletePhase
	deleteForce    bool
	deleting       bool
	portForwarding bool
	statusMsg      string
	err            error
	quitting       bool
}

// NewModel creates an initial TUI model.
func NewModel(opts Options) Model {
	return Model{
		opts:             opts,
		detailFormat:     DetailJSON,
		detailSerialized: DetailJSON,
		secsLeft:         opts.RefreshSecs,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchCmd(),
		m.contextCmd(),
		m.spinnerTick(),
		m.dataTick(),
	)
}

func (m Model) spinnerTick() tea.Cmd {
	return tea.Tick(spinnerInterval, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m Model) dataTick() tea.Cmd {
	return tea.Tick(time.Duration(m.opts.RefreshSecs)*time.Second, func(time.Time) tea.Msg {
		return dataTickMsg{}
	})
}

func (m Model) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.opts.Fetcher()
		if err != nil {
			return errMsg{err: err}
		}
		return entriesFetchedMsg{entries: entries}
	}
}

func (m Model) contextCmd() tea.Cmd {
	if m.opts.ContextProvider == nil {
		return nil
	}
	provider := m.opts.ContextProvider
	return func() tea.Msg {
		return contextRefreshedMsg{context: provider()}
	}
}

type entriesFetchedMsg struct {
	entries []ClusterEntry
}

type contextRefreshedMsg struct {
	context ContextInfo
}

type errMsg struct{ err error }

type patchResultMsg struct {
	info string
	err  error
}

type portForwardResultMsg struct {
	info string
	err  error
}

type deleteResultMsg struct {
	info string
	err  error
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case entriesFetchedMsg:
		m.snapshot = BuildSnapshot(msg.entries, m.tick, m.opts.RefreshSecs, m.opts.NoColor)
		if m.selected >= len(m.snapshot.Rows) {
			if len(m.snapshot.Rows) > 0 {
				m.selected = len(m.snapshot.Rows) - 1
			} else {
				m.selected = 0
			}
		}
		if m.selected < 0 {
			m.selected = 0
		}
		m.nextRefresh = time.Now().Add(time.Duration(m.opts.RefreshSecs) * time.Second)
		m.secsLeft = m.opts.RefreshSecs
		m.clampScroll()
		if m.viewMode == ViewFilter {
			m.refreshFilterTable()
		}
		if m.statusMsg == "Refreshing…" {
			m.statusMsg = ""
		}
		return m, nil

	case contextRefreshedMsg:
		m.context = msg.context
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case spinnerTickMsg:
		m.tick++
		m.secsLeft = secsUntil(m.nextRefresh)
		return m, m.spinnerTick()

	case dataTickMsg:
		return m, tea.Batch(m.fetchCmd(), m.contextCmd(), m.dataTick())

	case patchResultMsg:
		m.patching = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("[ERROR] %v", msg.err)
		} else {
			m.statusMsg = msg.info
		}
		return m, tea.Batch(m.fetchCmd(), m.contextCmd())

	case portForwardResultMsg:
		m.portForwarding = false
		if msg.err != nil {
			m.statusMsg = msg.info
			if msg.info == "" {
				m.statusMsg = fmt.Sprintf("[ERROR] %v", msg.err)
			}
		} else {
			m.statusMsg = msg.info
		}
		return m, m.contextCmd()

	case deleteResultMsg:
		m.deleting = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("[ERROR] %v", msg.err)
		} else {
			m.statusMsg = msg.info
		}
		return m, tea.Batch(m.fetchCmd(), m.contextCmd())

	case tea.KeyMsg:
		return m.updateKey(msg)
	}

	return m, nil
}

func (m Model) updateKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.patching || m.portForwarding || m.deleting {
		return m, nil
	}
	if m.deletePhase != DeletePhaseNone {
		return m.updateDeleteKey(msg)
	}
	if m.patchMode {
		return m.updatePatchKey(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "esc", "backspace":
		if m.detailOpen {
			m.closeDetail()
			return m, nil
		}
		if m.viewMode == ViewPicker || m.viewMode == ViewFilter {
			m.backToList()
			return m, nil
		}

	case "up", "k":
		m.statusMsg = ""
		return m.handleUp()

	case "down", "j":
		m.statusMsg = ""
		return m.handleDown()

	case "pgup":
		return m.handlePageUp()

	case "pgdown":
		return m.handlePageDown()

	case "enter":
		if m.viewMode == ViewPicker {
			m.confirmPicker()
			return m, nil
		}
		if m.viewMode == ViewList && !m.detailOpen && len(m.snapshot.Rows) > 0 {
			m.openDetail()
		}

	case "d":
		if m.opts.Deleter != nil && m.viewMode == ViewList && !m.detailOpen && len(m.snapshot.Rows) > 0 {
			m.deletePhase = DeletePhaseConfirm
			m.deleteForce = m.selectedResourceIsDeleted()
			m.statusMsg = ""
		}

	case "v":
		if m.detailOpen {
			if m.detailFormat == DetailOverview {
				m.detailFormat = m.detailSerialized
			} else {
				m.detailSerialized = m.detailFormat
				m.detailFormat = DetailOverview
			}
			m.detailScroll = 0
		}

	case "y":
		if m.detailOpen && m.detailFormat != DetailOverview {
			if m.detailFormat == DetailYAML {
				m.detailFormat = DetailJSON
			} else {
				m.detailFormat = DetailYAML
			}
			m.detailSerialized = m.detailFormat
			m.detailScroll = 0
		}

	case "r":
		if m.detailOpen || m.viewMode == ViewFilter {
			return m.runRefresh()
		}

	case "s":
		if m.viewMode == ViewList && !m.detailOpen && len(m.snapshot.Rows) > 0 {
			m.openPicker(FilterByCondition)
		}

	case "a":
		if m.viewMode == ViewList && !m.detailOpen && len(m.snapshot.Rows) > 0 {
			m.openPicker(FilterByAdapter)
		}

	case "c":
		if m.opts.Patcher != nil && len(m.snapshot.Rows) > 0 && m.viewMode == ViewList && !m.detailOpen {
			m.patchMode = true
			m.statusMsg = ""
		}

	case "p":
		if m.opts.PortForwardToggler != nil && m.viewMode == ViewList && !m.detailOpen {
			return m.runPortForwardToggle()
		}
	}

	return m, nil
}

func (m Model) handleUp() (Model, tea.Cmd) {
	switch {
	case m.detailOpen:
		m.detailScroll--
		m.clampScroll()
	case m.viewMode == ViewPicker:
		if m.pickerSelected > 0 {
			m.pickerSelected--
		}
	case m.viewMode == ViewFilter:
		if m.filterScroll > 0 {
			m.filterScroll--
		}
	case m.selected > 0:
		m.selected--
		m.ensureRowVisible()
	}
	return m, nil
}

func (m Model) handleDown() (Model, tea.Cmd) {
	switch {
	case m.detailOpen:
		m.detailScroll++
		m.clampScroll()
	case m.viewMode == ViewPicker:
		if m.pickerSelected < len(m.pickerItems)-1 {
			m.pickerSelected++
		}
	case m.viewMode == ViewFilter:
		maxScroll := len(m.filterRows) - m.mainPanelHeight()
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.filterScroll < maxScroll {
			m.filterScroll++
		}
	case m.selected < len(m.snapshot.Rows)-1:
		m.selected++
		m.ensureRowVisible()
	}
	return m, nil
}

func (m Model) handlePageUp() (Model, tea.Cmd) {
	switch {
	case m.detailOpen:
		m.detailScroll -= 10
		m.clampScroll()
	case m.viewMode == ViewFilter:
		m.filterScroll -= m.mainPanelHeight()
		if m.filterScroll < 0 {
			m.filterScroll = 0
		}
	default:
		m.tableOffsetY -= m.mainPanelHeight()
		if m.tableOffsetY < 0 {
			m.tableOffsetY = 0
		}
	}
	return m, nil
}

func (m Model) handlePageDown() (Model, tea.Cmd) {
	switch {
	case m.detailOpen:
		m.detailScroll += 10
		m.clampScroll()
	case m.viewMode == ViewFilter:
		maxScroll := len(m.filterRows) - m.mainPanelHeight()
		if maxScroll < 0 {
			maxScroll = 0
		}
		m.filterScroll += m.mainPanelHeight()
		if m.filterScroll > maxScroll {
			m.filterScroll = maxScroll
		}
	default:
		m.tableOffsetY += m.mainPanelHeight()
		m.ensureRowVisible()
	}
	return m, nil
}

func (m *Model) openPicker(kind FilterKind) {
	_, _, statuses, ok := m.snapshot.SelectedResource(m.selected)
	if !ok || len(statuses) == 0 {
		m.statusMsg = "[WARN] No adapter statuses for selected resource"
		return
	}

	var items []string
	switch kind {
	case FilterByCondition:
		items = collectConditionTypes(statuses)
	case FilterByAdapter:
		items = collectAdapterNames(statuses)
	}
	if len(items) == 0 {
		m.statusMsg = "[WARN] Nothing to filter"
		return
	}

	m.viewMode = ViewPicker
	m.pickerKind = kind
	m.pickerItems = items
	m.pickerSelected = 0
	m.statusMsg = ""
}

func (m *Model) confirmPicker() {
	if m.pickerSelected < 0 || m.pickerSelected >= len(m.pickerItems) {
		return
	}
	m.filterKind = m.pickerKind
	m.filterKey = m.pickerItems[m.pickerSelected]
	m.viewMode = ViewFilter
	m.filterScroll = 0
	m.pickerItems = nil
	m.refreshFilterTable()
}

func (m *Model) refreshFilterTable() {
	_, _, statuses, ok := m.snapshot.SelectedResource(m.selected)
	if !ok {
		m.filterHeaders = nil
		m.filterRows = nil
		return
	}
	switch m.filterKind {
	case FilterByCondition:
		m.filterHeaders, m.filterRows = buildConditionTypeTable(statuses, m.filterKey, m.opts.NoColor)
	case FilterByAdapter:
		m.filterHeaders, m.filterRows = buildAdapterConditionsTable(statuses, m.filterKey, m.opts.NoColor)
	}
}

func (m *Model) backToList() {
	m.viewMode = ViewList
	m.pickerItems = nil
	m.pickerSelected = 0
	m.filterHeaders = nil
	m.filterRows = nil
	m.filterKey = ""
	m.filterScroll = 0
}

func (m *Model) openDetail() {
	m.detailOpen = true
	m.detailScroll = 0
	m.detailFormat = DetailJSON
	m.detailSerialized = DetailJSON
}

func (m *Model) closeDetail() {
	m.detailOpen = false
	m.detailScroll = 0
}

func (m Model) updatePatchKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "c":
		m.patchMode = false
	case "s":
		return m.runPatch("spec")
	case "l":
		return m.runPatch("labels")
	}
	return m, nil
}

func (m Model) updateDeleteKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "backspace":
		m.deletePhase = DeletePhaseNone
		m.deleteForce = false
	case "y":
		if m.deletePhase == DeletePhaseConfirm {
			return m.runDelete(m.deleteForce)
		}
	}
	return m, nil
}

func (m Model) runDelete(force bool) (Model, tea.Cmd) {
	m.deletePhase = DeletePhaseNone
	m.deleteForce = false
	target, ok := m.selectedPatchTarget()
	if !ok || m.opts.Deleter == nil {
		return m, nil
	}
	m.deleting = true
	if force {
		m.statusMsg = "Force-deleting…"
	} else {
		m.statusMsg = "Deleting…"
	}
	deleter := m.opts.Deleter
	return m, func() tea.Msg {
		info, err := deleter(target, force)
		return deleteResultMsg{info: info, err: err}
	}
}

func (m Model) selectedResourceIsDeleted() bool {
	cl, np, _, ok := m.snapshot.SelectedResource(m.selected)
	if !ok {
		return false
	}
	if np != nil {
		return np.DeletedTime != ""
	}
	return cl.DeletedTime != ""
}

func (m Model) deleteTargetLabel() string {
	cl, np, _, ok := m.snapshot.SelectedResource(m.selected)
	if !ok {
		return "resource"
	}
	if np != nil {
		return fmt.Sprintf("nodepool %q", np.Name)
	}
	return fmt.Sprintf("cluster %q", cl.Name)
}

func (m Model) runRefresh() (Model, tea.Cmd) {
	m.statusMsg = "Refreshing…"
	return m, tea.Batch(m.fetchCmd(), m.contextCmd())
}

func (m Model) runPortForwardToggle() (Model, tea.Cmd) {
	m.portForwarding = true
	m.statusMsg = "Port-forwards…"
	toggler := m.opts.PortForwardToggler
	return m, func() tea.Msg {
		info, err := toggler()
		return portForwardResultMsg{info: info, err: err}
	}
}

func (m Model) runPatch(section string) (Model, tea.Cmd) {
	m.patchMode = false
	target, ok := m.selectedPatchTarget()
	if !ok || m.opts.Patcher == nil {
		return m, nil
	}
	m.patching = true
	m.statusMsg = "Patching…"
	patcher := m.opts.Patcher
	return m, func() tea.Msg {
		info, err := patcher(target, section)
		return patchResultMsg{info: info, err: err}
	}
}

func (m Model) selectedPatchTarget() (PatchTarget, bool) {
	cl, np, _, ok := m.snapshot.SelectedResource(m.selected)
	if !ok || cl == nil {
		return PatchTarget{}, false
	}
	if np != nil {
		return PatchTarget{IsNodePool: true, ClusterID: cl.ID, NodePoolID: np.ID}, true
	}
	return PatchTarget{ClusterID: cl.ID}, true
}

func (m *Model) clampScroll() {
	content := m.detailContent()
	lines := strings.Split(content, "\n")
	maxScroll := len(lines) - m.mainPanelHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.detailScroll < 0 {
		m.detailScroll = 0
	}
	if m.detailScroll > maxScroll {
		m.detailScroll = maxScroll
	}
}

func (m *Model) ensureRowVisible() {
	visible := m.mainPanelHeight()
	if m.selected < m.tableOffsetY {
		m.tableOffsetY = m.selected
	}
	if m.selected >= m.tableOffsetY+visible {
		m.tableOffsetY = m.selected - visible + 1
	}
}

func (m Model) detailContent() string {
	cl, np, statuses, ok := m.snapshot.SelectedResource(m.selected)
	if !ok {
		return ""
	}
	return RenderDetail(cl, np, statuses, m.detailFormat, m.opts.NoColor)
}

func (m Model) contentHeight() int {
	h := m.height - m.headerLineCount() - 1 - 2 // separator + footer + spacing
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) detailFormatLabel() string {
	switch m.detailFormat {
	case DetailYAML:
		return "yaml"
	case DetailOverview:
		return "overview"
	default:
		return "json"
	}
}

func (m Model) mainPanelHeight() int {
	return m.contentHeight()
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.err != nil {
		return fmt.Sprintf("[ERROR] %v\n", m.err)
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderHeaderSeparator())
	b.WriteString("\n")

	switch {
	case m.detailOpen:
		b.WriteString(m.viewDetailPanel())
	case m.viewMode == ViewPicker:
		b.WriteString(m.viewPicker())
	case m.viewMode == ViewFilter:
		b.WriteString(m.viewFilterTable())
	default:
		b.WriteString(m.viewTable(m.width))
	}

	b.WriteString("\n")
	b.WriteString(m.viewFooter())

	out := b.String()
	if m.deletePhase != DeletePhaseNone {
		out = overlayModal(out, m.renderDeleteModal(), m.width, m.height)
	}
	return out
}

func (m Model) viewTable(_ int) string {
	if len(m.snapshot.Headers) == 0 {
		return lipgloss.NewStyle().Italic(true).Render("(no data)")
	}

	ft := formatMainTable(m.snapshot.Headers, m.snapshot.Rows, 0)
	return renderTableBlock(ft, m.selected, m.tableOffsetY, m.mainPanelHeight(), m.width, m.opts.NoColor)
}

func (m Model) viewFilterTable() string {
	if len(m.filterHeaders) == 0 {
		return lipgloss.NewStyle().Italic(true).Render("(no matching conditions)")
	}
	ft := formatMainTable(m.filterHeaders, m.filterRows, 0)
	return renderTableBlock(ft, -1, m.filterScroll, m.mainPanelHeight(), m.width, m.opts.NoColor)
}

func (m Model) viewDetailPanel() string {
	content := m.detailContent()
	lines := strings.Split(content, "\n")

	start := m.detailScroll
	end := start + m.mainPanelHeight()
	if end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		start = len(lines)
	}

	var visible []string
	for i := start; i < end; i++ {
		visible = append(visible, lines[i])
	}

	body := strings.Join(visible, "\n")
	border := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(m.width - 2)
	return border.Render(body)
}

func (m Model) viewFooter() string {
	if m.patchMode {
		return lipgloss.NewStyle().Faint(true).Render("Patch spec (s) or labels (l) · Esc cancel")
	}
	if m.detailOpen {
		return lipgloss.NewStyle().Faint(true).Render("↑↓ scroll · V overview · y json/yaml · r refresh · Esc back · q quit")
	}
	if m.viewMode == ViewPicker {
		return lipgloss.NewStyle().Faint(true).Render("↑↓ select · Enter confirm · Esc cancel · q quit")
	}
	if m.viewMode == ViewFilter {
		return lipgloss.NewStyle().Faint(true).Render("↑↓ scroll · r refresh · Esc back · q quit")
	}
	return lipgloss.NewStyle().Faint(true).Render("↑↓ navigate · Enter describe · d delete · s condition · a adapter · c patch · p port-forwards · q quit")
}

// Run starts the TUI event loop.
func Run(opts Options) error {
	if opts.RefreshSecs < 1 {
		opts.RefreshSecs = 1
	}
	m := NewModel(opts)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func secsUntil(t time.Time) int {
	d := time.Until(t)
	if d <= 0 {
		return 0
	}
	return int((d + time.Second - 1) / time.Second)
}
