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
	if !EnsureKeyBindings(&settings) {
		t.Fatal("expected missing default actions to be added")
	}
	if got, want := settings.KeyBindings["refresh"], []string{"ctrl+r"}; !sameStrings(got, want) {
		t.Fatalf("refresh = %v, want override %v", got, want)
	}
	if len(settings.KeyBindings["scroll_down"]) == 0 {
		t.Fatal("scroll_down default was not added")
	}
	if settings.KeyBindingsVersion != keyBindingsVersion {
		t.Fatalf("keybindings version = %d, want %d", settings.KeyBindingsVersion, keyBindingsVersion)
	}
}

func TestEnsureKeyBindingsMigratesStaleDefaults(t *testing.T) {
	settings := ConfigSettings{KeyBindings: map[string][]string{
		"focus_notifications": {"N"},
		"next_page":           {"n", "right"},
	}}
	if !EnsureKeyBindings(&settings) {
		t.Fatal("expected stale defaults to be migrated")
	}
	if got, want := settings.KeyBindings["focus_notifications"], []string{"N", "n"}; !sameStrings(got, want) {
		t.Fatalf("focus_notifications = %v, want %v", got, want)
	}
	if got, want := settings.KeyBindings["next_page"], []string{"O", "o", "right"}; !sameStrings(got, want) {
		t.Fatalf("next_page = %v, want %v", got, want)
	}
}

func TestEnsureKeyBindingsPreservesUserOverridesDuringMigration(t *testing.T) {
	settings := ConfigSettings{KeyBindings: map[string][]string{
		"focus_notifications": {"f"},
		"next_page":           {"n", "right"},
		"jukebox_next":        {"up", "k"},
	}}
	if !EnsureKeyBindings(&settings) {
		t.Fatal("expected a migrated next_page to be reported")
	}
	if got, want := settings.KeyBindings["focus_notifications"], []string{"f"}; !sameStrings(got, want) {
		t.Fatalf("focus_notifications = %v, want preserved override %v", got, want)
	}
	if got, want := settings.KeyBindings["jukebox_next"], []string{"up", "k"}; !sameStrings(got, want) {
		t.Fatalf("jukebox_next = %v, want preserved override %v", got, want)
	}
	if got, want := settings.KeyBindings["next_page"], []string{"O", "o", "right"}; !sameStrings(got, want) {
		t.Fatalf("next_page = %v, want migrated default %v", got, want)
	}
}

func TestEnsureKeyBindingsIdempotentAtCurrentVersion(t *testing.T) {
	settings := ConfigSettings{
		KeyBindingsVersion: keyBindingsVersion,
		KeyBindings:        DefaultKeyBindings(),
	}
	if EnsureKeyBindings(&settings) {
		t.Fatal("current-version config should not be rewritten")
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
