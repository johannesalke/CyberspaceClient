package client

import "testing"

func TestResolveKeyBindingsPreservesDefaultsAndOverrides(t *testing.T) {
	bindings := ResolveKeyBindings(map[string][]string{
		"scroll_up": {"w"},
		"next_page": {},
	})

	if got, want := bindings["scroll_up"], []string{"w"}; !sameStrings(got, want) {
		t.Fatalf("scroll_up = %v, want %v", got, want)
	}
	if got := bindings["next_page"]; len(got) != 0 {
		t.Fatalf("next_page = %v, want disabled action", got)
	}
	if got, want := bindings["quit"], []string{"q", "ctrl+c"}; !sameStrings(got, want) {
		t.Fatalf("quit = %v, want default %v", got, want)
	}
}

func TestEnsureKeyBindingsAddsMissingActionsWithoutReplacingOverrides(t *testing.T) {
	settings := ConfigSettings{KeyBindings: map[string][]string{"refresh": {"ctrl+r"}}}
	if !ensureKeyBindings(&settings) {
		t.Fatal("expected missing default actions to be added")
	}
	if got, want := settings.KeyBindings["refresh"], []string{"ctrl+r"}; !sameStrings(got, want) {
		t.Fatalf("refresh = %v, want override %v", got, want)
	}
	if len(settings.KeyBindings["scroll_down"]) == 0 {
		t.Fatal("scroll_down default was not added")
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
