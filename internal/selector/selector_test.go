package selector

import "testing"

// Compile-time check: FuzzySelector must implement Selector.
var _ Selector = FuzzySelector{}

// Compile-time check: FuzzyPreviewSelector must implement PreviewSelector.
var _ PreviewSelector = FuzzyPreviewSelector{}

func TestItemFields(t *testing.T) {
	item := Item{ID: "abc-123", Name: "my-cluster"}
	if item.ID != "abc-123" {
		t.Errorf("ID: got %q", item.ID)
	}
	if item.Name != "my-cluster" {
		t.Errorf("Name: got %q", item.Name)
	}
}
