package structural

import (
	"fmt"
	"os"
	"sort"
)

// IdeaGraph is the intermediate representation passed to all structural exporters.
type IdeaGraph struct {
	Idea       map[string]any
	Identities []map[string]any
	Places     []map[string]any
	Things     []map[string]any
	Docs       []map[string]any
	PluginData map[string]any // reserved — populated when plugin registry exists
}

// Exporter serialises an IdeaGraph to a structural format.
type Exporter interface {
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

// Export is the controller entry point. It normalises the manifest, checks the
// output path is safe, resolves the idea graph from the API, and hands it to
// the driver for serialisation.
func Export(client Fetcher, driver Exporter, m Manifest, ideaID, output string) error {
	m = m.Resolve()

	if output != "" {
		if err := checkOutputDir(output); err != nil {
			return err
		}
	}

	graph, err := Resolve(client, ideaID, m)
	if err != nil {
		return translateAPIError(err)
	}

	return driver.Export(graph, output)
}

// checkOutputDir returns an error if the output directory exists and is non-empty.
func checkOutputDir(output string) error {
	entries, err := os.ReadDir(output)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("output directory %q already exists and is not empty (remove it first)", output)
	}
	return nil
}

// translateAPIError converts raw HTTP errors from the resolver into
// user-facing messages for known platform responses.
func translateAPIError(err error) error {
	if err == nil {
		return nil
	}
	// TODO: inspect status codes for 403/402/429 once the API client exposes them.
	return err
}

func strField(row map[string]any, key string) string {
	if v, ok := row[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
