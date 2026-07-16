package structural

// Fetcher is the subset of the API client used by the resolver.
type Fetcher interface {
	Get(path string, out any) error
}

// Resolve fetches all entity data declared by the manifest and returns an IdeaGraph.
// Entity paths mirror the CLI's actual API structure:
//   - persons are nested under the idea: /idea/{id}/persons/
//   - identities are nested under each person: /person/{personID}/identities/
//   - places, things, and docs are nested under the idea directly
func Resolve(client Fetcher, ideaID string, m Manifest) (IdeaGraph, error) {
	m = m.Resolve()

	graph := IdeaGraph{
		PluginData: map[string]any{},
	}

	var idea map[string]any
	if err := client.Get("/idea/"+ideaID+"/", &idea); err != nil {
		return graph, err
	}
	graph.Idea = idea

	if m.Identities {
		var persons []map[string]any
		if err := client.Get("/idea/"+ideaID+"/persons/", &persons); err != nil {
			return graph, err
		}

		for _, person := range persons {
			personID := strField(person, "id")

			var items []map[string]any
			if err := client.Get("/person/"+personID+"/identities/", &items); err != nil {
				return graph, err
			}
			graph.Identities = append(graph.Identities, items...)
		}
	}

	if m.Places {
		var items []map[string]any
		if err := client.Get("/idea/"+ideaID+"/places/", &items); err != nil {
			return graph, err
		}
		graph.Places = items
	}

	if m.Things {
		var items []map[string]any
		if err := client.Get("/idea/"+ideaID+"/things/", &items); err != nil {
			return graph, err
		}
		graph.Things = items
	}

	if m.Docs {
		var items []map[string]any
		if err := client.Get("/idea/"+ideaID+"/documents/", &items); err != nil {
			return graph, err
		}
		graph.Docs = items
	}

	return graph, nil
}
