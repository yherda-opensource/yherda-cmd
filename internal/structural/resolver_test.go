package structural

import (
	"fmt"
	"testing"
)

type mockFetcher struct {
	calls    []string
	roleIDs  []string
	arcIDs   []string
}

func (m *mockFetcher) Get(path string, out any) error {
	m.calls = append(m.calls, path)
	switch v := out.(type) {
	case *[]map[string]any:
		// Return synthetic roles/arcs so the resolver can traverse them.
		if path == "/storyline/42/roles/" {
			*v = rolesFor(m.roleIDs)
		} else if isRoleArcsPath(path, m.roleIDs) {
			*v = arcsFor(m.arcIDs)
		} else {
			*v = nil
		}
	case *map[string]any:
		*v = map[string]any{"id": "42", "name": "test"}
	}
	return nil
}

func rolesFor(ids []string) []map[string]any {
	roles := make([]map[string]any, len(ids))
	for i, id := range ids {
		roles[i] = map[string]any{"id": id}
	}
	return roles
}

func arcsFor(ids []string) []map[string]any {
	arcs := make([]map[string]any, len(ids))
	for i, id := range ids {
		arcs[i] = map[string]any{"id": id}
	}
	return arcs
}

func isRoleArcsPath(path string, roleIDs []string) bool {
	for _, id := range roleIDs {
		if path == "/role/"+id+"/arcs/" {
			return true
		}
	}
	return false
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
	for _, unexpected := range []string{
		"/storyline/42/roles/",
		"/storyline/42/identities/",
		"/storyline/42/arcs/",
		"/storyline/42/beats/",
		"/storyline/42/things/",
		"/storyline/42/documents/",
	} {
		if hasCall(f.calls, unexpected) {
			t.Errorf("unexpected call to %s", unexpected)
		}
	}
}

func TestResolver_BeatsTriggersFullIdentityChain(t *testing.T) {
	f := &mockFetcher{roleIDs: []string{"r1"}, arcIDs: []string{"a1"}}
	_, err := Resolve(f, "42", Manifest{Beats: true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCall(f.calls, "/storyline/42/roles/") {
		t.Error("expected call to /storyline/42/roles/")
	}
	if !hasCall(f.calls, "/role/r1/identities/") {
		t.Error("expected call to /role/r1/identities/")
	}
	if !hasCall(f.calls, "/role/r1/arcs/") {
		t.Error("expected call to /role/r1/arcs/")
	}
	if !hasCall(f.calls, "/arc/a1/beats/") {
		t.Error("expected call to /arc/a1/beats/")
	}
}

func TestResolver_ArcsPerRole(t *testing.T) {
	f := &mockFetcher{roleIDs: []string{"r1", "r2"}, arcIDs: []string{"a1"}}
	_, err := Resolve(f, "42", Manifest{Arcs: true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCall(f.calls, "/role/r1/arcs/") {
		t.Error("expected call to /role/r1/arcs/")
	}
	if !hasCall(f.calls, "/role/r2/arcs/") {
		t.Error("expected call to /role/r2/arcs/")
	}
	// Beats should not be fetched.
	for _, c := range f.calls {
		if len(c) > 4 && c[len(c)-7:] == "/beats/" {
			t.Errorf("unexpected beats call: %s", c)
		}
	}
}

func TestResolver_IdentitiesNotFetchedForPlacesOnly(t *testing.T) {
	f := &mockFetcher{}
	_, err := Resolve(f, "42", Manifest{Places: true, Things: true})
	if err != nil {
		t.Fatal(err)
	}
	if hasCall(f.calls, "/storyline/42/roles/") {
		t.Error("roles should not be fetched when only Places/Things requested")
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
