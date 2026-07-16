package structural

import (
	"fmt"
	"testing"
)

// stubPoster records POST calls in order for assertion.
type stubPoster struct {
	calls  []postCall
	nextID int
	errOn  string // if set, return an error when path contains this substring
}

type postCall struct {
	path string
	body any
}

func (s *stubPoster) Post(path string, body any, out any) error {
	if s.errOn != "" && len(path) >= len(s.errOn) {
		for i := 0; i <= len(path)-len(s.errOn); i++ {
			if path[i:i+len(s.errOn)] == s.errOn {
				return fmt.Errorf("stub error on %s", path)
			}
		}
	}
	s.calls = append(s.calls, postCall{path: path, body: body})
	s.nextID++
	if out != nil {
		if m, ok := out.(*map[string]any); ok {
			*m = map[string]any{"id": fmt.Sprintf("%d", s.nextID)}
		}
	}
	return nil
}

func (s *stubPoster) pathsContaining(sub string) []string {
	var found []string
	for _, c := range s.calls {
		if contains(c.path, sub) {
			found = append(found, c.path)
		}
	}
	return found
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}

func TestMarshal_DependencyOrder(t *testing.T) {
	graph := IdeaGraph{
		PluginData: map[string]any{},
		Identities: []map[string]any{
			{"_id": "ident-1", "name": "Alice"},
		},
		Places: []map[string]any{
			{"name": "The Forest"},
		},
		Things: []map[string]any{
			{"name": "Magic sword"},
		},
		Docs: []map[string]any{
			{"title": "World notes", "body": "Some notes"},
		},
	}

	poster := &stubPoster{}
	if err := Marshal(poster, graph, "42"); err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify something was posted to each entity type.
	if len(poster.pathsContaining("persons")) == 0 {
		t.Error("expected POST to /persons/")
	}
	if len(poster.pathsContaining("identities")) == 0 {
		t.Error("expected POST to /identities/")
	}
	if len(poster.pathsContaining("places")) == 0 {
		t.Error("expected POST to /places/")
	}
	if len(poster.pathsContaining("things")) == 0 {
		t.Error("expected POST to /things/")
	}
	if len(poster.pathsContaining("documents")) == 0 {
		t.Error("expected POST to /documents/")
	}

	// Verify ordering: person before identity.
	personIdx, identityIdx := -1, -1
	for i, c := range poster.calls {
		if contains(c.path, "persons") && personIdx < 0 {
			personIdx = i
		}
		if contains(c.path, "identities") && identityIdx < 0 {
			identityIdx = i
		}
	}
	if personIdx >= identityIdx {
		t.Errorf("person POST (idx %d) must come before identity POST (idx %d)", personIdx, identityIdx)
	}
}

func TestMarshal_EmptyGraph(t *testing.T) {
	graph := IdeaGraph{PluginData: map[string]any{}}
	poster := &stubPoster{}
	if err := Marshal(poster, graph, "42"); err != nil {
		t.Fatalf("Marshal of empty graph should not error: %v", err)
	}
	if len(poster.calls) != 0 {
		t.Errorf("expected no API calls for empty graph, got %d", len(poster.calls))
	}
}
