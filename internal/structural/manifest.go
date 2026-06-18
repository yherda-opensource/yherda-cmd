package structural

// Manifest declares which entity types a format driver needs.
// Dependency rules are enforced here: Beats implies Arcs implies Identities.
type Manifest struct {
	Identities bool
	Arcs       bool
	Beats      bool
	Places     bool
	Things     bool
	Docs       bool
	Plugins    []string // reserved — ignored until plugin registry exists
}

// Resolve enforces dependency rules and returns the normalised manifest.
func (m Manifest) Resolve() Manifest {
	if m.Beats {
		m.Arcs = true
	}
	if m.Arcs {
		m.Identities = true
	}
	return m
}
