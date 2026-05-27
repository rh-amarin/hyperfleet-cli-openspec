package tui

// ContextInfo holds environment and connectivity info shown in the TUI header.
type ContextInfo struct {
	ActiveEnv        string
	APIURL           string
	KubeContext      string
	AutoPortForward  bool
	PortForwards     []PortForwardLine
}

// PortForwardLine is one port-forward entry in the header.
type PortForwardLine struct {
	Name      string
	LocalPort int
	Connected bool
}

// ContextProvider returns current config/connectivity snapshot for the header.
type ContextProvider func() ContextInfo
