package structural

// Fetcher is the subset of the API client used by the resolver.
type Fetcher interface {
	Get(path string, out any) error
}

// Resolve fetches all entity data declared by the manifest and returns an IdeaGraph.
func Resolve(client Fetcher, ideaID string, m Manifest) (IdeaGraph, error) {
	m = m.Resolve()

	graph := IdeaGraph{
		PluginData: map[string]any{},
	}

	var idea map[string]any
	if err := client.Get("/storyline/"+ideaID+"/", &idea); err != nil {
		return graph, err
	}
	graph.Idea = idea

	if m.Identities {
		var items []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/identities/", &items); err != nil {
			return graph, err
		}
		graph.Identities = items
	}

	if m.Arcs {
		var items []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/arcs/", &items); err != nil {
			return graph, err
		}
		graph.Arcs = items
	}

	if m.Beats {
		var items []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/beats/", &items); err != nil {
			return graph, err
		}
		graph.Beats = items
	}

	if m.Places {
		var items []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/places/", &items); err != nil {
			return graph, err
		}
		graph.Places = items
	}

	if m.Things {
		var items []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/things/", &items); err != nil {
			return graph, err
		}
		graph.Things = items
	}

	if m.Docs {
		var items []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/documents/", &items); err != nil {
			return graph, err
		}
		graph.Docs = items
	}

	return graph, nil
}
