package structural

// Manifest declares which entity types a format driver needs.
type Manifest struct {
	Identities bool
	Places     bool
	Things     bool
	Docs       bool
	Plugins    []string // reserved — ignored until plugin registry exists
}

// Resolve normalises the manifest. Reserved for future dependency rules.
func (m Manifest) Resolve() Manifest {
	return m
}
