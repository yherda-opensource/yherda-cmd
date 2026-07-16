package structural

import "testing"

func TestManifest_IndependentEntities(t *testing.T) {
	m := Manifest{Places: true, Things: true}.Resolve()
	if m.Identities {
		t.Error("Places/Things should not imply Identities")
	}
	if !m.Places || !m.Things {
		t.Error("expected Places and Things to remain true")
	}
}
