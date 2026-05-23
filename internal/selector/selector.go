package selector

import (
	"fmt"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// Item is a selectable row displayed as "<Name>  <ID>" in the picker.
type Item struct {
	ID   string
	Name string
}

// Selector picks one item from a list.
// Returns index -1 and nil error when the user aborts (Esc / Ctrl+C).
type Selector interface {
	Select(items []Item) (int, error)
}

// FuzzySelector is the live implementation backed by go-fuzzyfinder.
type FuzzySelector struct{}

func (FuzzySelector) Select(items []Item) (int, error) {
	idx, err := fuzzyfinder.Find(items, func(i int) string {
		return fmt.Sprintf("%-40s  %s", items[i].Name, items[i].ID)
	})
	if err == fuzzyfinder.ErrAbort {
		return -1, nil
	}
	return idx, err
}

// PreviewSelector picks one item from a list and shows a preview panel on the right.
// Returns index -1 and nil error when the user aborts (Esc / Ctrl+C).
type PreviewSelector interface {
	SelectWithPreview(items []Item, preview func(i int) string) (int, error)
}

// FuzzyPreviewSelector is the live implementation backed by go-fuzzyfinder.
// The preview panel is rendered on the right half of the terminal.
type FuzzyPreviewSelector struct{}

func (FuzzyPreviewSelector) SelectWithPreview(items []Item, preview func(i int) string) (int, error) {
	idx, err := fuzzyfinder.Find(
		items,
		func(i int) string { return items[i].Name },
		fuzzyfinder.WithPreviewWindow(func(i, _, _ int) string {
			if i == -1 {
				return ""
			}
			return preview(i)
		}),
	)
	if err == fuzzyfinder.ErrAbort {
		return -1, nil
	}
	return idx, err
}
