package structural

import "testing"

func TestManifest_BeatsImpliesArcsImpliesIdentities(t *testing.T) {
	m := Manifest{Beats: true}.Resolve()
	if !m.Arcs {
		t.Error("expected Arcs=true when Beats=true")
	}
	if !m.Identities {
		t.Error("expected Identities=true when Beats=true")
	}
}

func TestManifest_ArcsImpliesIdentities(t *testing.T) {
	m := Manifest{Arcs: true}.Resolve()
	if !m.Identities {
		t.Error("expected Identities=true when Arcs=true")
	}
	if m.Beats {
		t.Error("expected Beats=false")
	}
}

func TestManifest_IndependentEntities(t *testing.T) {
	m := Manifest{Places: true, Things: true}.Resolve()
	if m.Identities || m.Arcs || m.Beats {
		t.Error("Places/Things should not imply Identities, Arcs, or Beats")
	}
}
