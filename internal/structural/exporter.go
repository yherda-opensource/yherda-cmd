package structural

import (
	"fmt"
	"sort"
)

// IdeaGraph is the intermediate representation passed to all structural exporters.
type IdeaGraph struct {
	Idea       map[string]any
	Identities []map[string]any
	Arcs       []map[string]any
	Beats      []map[string]any
	Places     []map[string]any
	Things     []map[string]any
	Docs       []map[string]any
	PluginData map[string]any // reserved — populated when plugin registry exists
}

// Exporter serialises an IdeaGraph to a structural format.
type Exporter interface {
	Manifest() Manifest
	Export(graph IdeaGraph, output string) error
	DefaultExt() string
}

var registry = map[string]Exporter{
	"obsidian": &ObsidianExporter{},
}

// Get returns the exporter for the given format name.
func Get(format string) (Exporter, bool) {
	e, ok := registry[format]
	return e, ok
}

// Formats returns a sorted list of registered format names.
func Formats() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func strField(row map[string]any, key string) string {
	if v, ok := row[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
