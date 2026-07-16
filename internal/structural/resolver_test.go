package structural

import (
	"fmt"
	"testing"
)

type mockFetcher struct {
	calls     []string
	personIDs []string
}

func (m *mockFetcher) Get(path string, out any) error {
	m.calls = append(m.calls, path)
	switch v := out.(type) {
	case *[]map[string]any:
		if path == "/storyline/42/persons/" {
			*v = personsFor(m.personIDs)
		} else {
			*v = nil
		}
	case *map[string]any:
		*v = map[string]any{"id": "42", "name": "test"}
	}
	return nil
}

func personsFor(ids []string) []map[string]any {
	persons := make([]map[string]any, len(ids))
	for i, id := range ids {
		persons[i] = map[string]any{"id": id}
	}
	return persons
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
		"/storyline/42/persons/",
		"/storyline/42/things/",
		"/storyline/42/documents/",
	} {
		if hasCall(f.calls, unexpected) {
			t.Errorf("unexpected call to %s", unexpected)
		}
	}
}

func TestResolver_IdentitiesFetchedPerPerson(t *testing.T) {
	f := &mockFetcher{personIDs: []string{"p1", "p2"}}
	_, err := Resolve(f, "42", Manifest{Identities: true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCall(f.calls, "/storyline/42/persons/") {
		t.Error("expected call to /storyline/42/persons/")
	}
	if !hasCall(f.calls, "/person/p1/identities/") {
		t.Error("expected call to /person/p1/identities/")
	}
	if !hasCall(f.calls, "/person/p2/identities/") {
		t.Error("expected call to /person/p2/identities/")
	}
}

func TestResolver_PersonsNotFetchedForPlacesOnly(t *testing.T) {
	f := &mockFetcher{}
	_, err := Resolve(f, "42", Manifest{Places: true, Things: true})
	if err != nil {
		t.Fatal(err)
	}
	if hasCall(f.calls, "/storyline/42/persons/") {
		t.Error("persons should not be fetched when only Places/Things requested")
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
