package structural

import (
	"fmt"
	"sort"
)

// Importer parses a source path and returns an IdeaGraph.
type Importer interface {
	Import(source string) (IdeaGraph, error)
}

var importerRegistry = map[string]Importer{
	"scriv": &ScrivImporter{},
}

// GetImporter returns the importer for the given format name.
func GetImporter(format string) (Importer, bool) {
	i, ok := importerRegistry[format]
	return i, ok
}

// ImportFormats returns a sorted list of registered import format names.
func ImportFormats() []string {
	names := make([]string, 0, len(importerRegistry))
	for k := range importerRegistry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// Import is the controller entry point. It runs the driver to produce an IdeaGraph,
// then either prints a dry-run summary or marshals the graph to the API.
func Import(client Poster, driver Importer, ideaID, source string, dryRun bool) error {
	graph, err := driver.Import(source)
	if err != nil {
		return err
	}

	if dryRun {
		printDryRun(graph)
		return nil
	}

	return Marshal(client, graph, ideaID)
}

func printDryRun(graph IdeaGraph) {
	unmapped := 0
	for _, doc := range graph.Docs {
		if strField(doc, "_unmapped") == "true" {
			unmapped++
		}
	}
	regularDocs := len(graph.Docs) - unmapped

	fmt.Println("Would create:")
	fmt.Printf("  %d identities\n", len(graph.Identities))
	fmt.Printf("  %d arcs\n", len(graph.Arcs))
	fmt.Printf("  %d beats\n", len(graph.Beats))
	fmt.Printf("  %d places\n", len(graph.Places))
	fmt.Printf("  %d things\n", len(graph.Things))
	if regularDocs > 0 {
		fmt.Printf("  %d idea documents\n", regularDocs)
	}
	if unmapped > 0 {
		fmt.Printf("  %d unmapped items (saved as idea documents)\n", unmapped)
	}
}
