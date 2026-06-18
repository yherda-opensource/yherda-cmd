package structural

import (
	"fmt"
	"testing"
)

type mockFetcher struct {
	calls []string
}

func (m *mockFetcher) Get(path string, out any) error {
	m.calls = append(m.calls, path)
	// Return empty slice or map so callers don't crash.
	switch v := out.(type) {
	case *[]map[string]any:
		*v = nil
	case *map[string]any:
		*v = map[string]any{"id": "1", "name": "test"}
	}
	return nil
}

func hasCall(calls []string, path string) bool {
	for _, c := range calls {
		if c == path {
			return true
		}
	}
	return false
}

func TestResolver_OnlyCallsDeclaredEndpoints(t *testing.T) {
	f := &mockFetcher{}
	_, err := Resolve(f, "42", Manifest{Places: true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCall(f.calls, "/storyline/42/") {
		t.Error("expected call to /storyline/42/")
	}
	if !hasCall(f.calls, "/storyline/42/places/") {
		t.Error("expected call to /storyline/42/places/")
	}
	for _, unexpected := range []string{"/storyline/42/identities/", "/storyline/42/arcs/", "/storyline/42/beats/", "/storyline/42/things/", "/storyline/42/documents/"} {
		if hasCall(f.calls, unexpected) {
			t.Errorf("unexpected call to %s", unexpected)
		}
	}
}

func TestResolver_BeatsTriggersFullIdentityChain(t *testing.T) {
	f := &mockFetcher{}
	_, err := Resolve(f, "42", Manifest{Beats: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"/storyline/42/identities/", "/storyline/42/arcs/", "/storyline/42/beats/"} {
		if !hasCall(f.calls, expected) {
			t.Errorf("expected call to %s", expected)
		}
	}
}

func TestResolver_PropagatesAPIError(t *testing.T) {
	f := &errorFetcher{err: fmt.Errorf("network error")}
	_, err := Resolve(f, "1", Manifest{})
	if err == nil {
		t.Error("expected error from fetcher, got nil")
	}
}

type errorFetcher struct{ err error }

func (e *errorFetcher) Get(_ string, _ any) error { return e.err }
