package structural

// Fetcher is the subset of the API client used by the resolver.
type Fetcher interface {
	Get(path string, out any) error
}

// Resolve fetches all entity data declared by the manifest and returns an IdeaGraph.
// Entity paths mirror the CLI's actual API structure:
//   - roles (persons) are nested under the idea: /storyline/{id}/roles/
//   - identities and arcs are nested under each role: /role/{roleID}/identities|arcs/
//   - beats are nested under each arc: /arc/{arcID}/beats/
//   - places, things, and docs are nested under the idea directly
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

	// Roles are always fetched when any of identities, arcs, or beats are needed.
	var roles []map[string]any
	if m.Identities || m.Arcs || m.Beats {
		if err := client.Get("/storyline/"+ideaID+"/roles/", &roles); err != nil {
			return graph, err
		}
	}

	for _, role := range roles {
		roleID := strField(role, "id")

		if m.Identities {
			var items []map[string]any
			if err := client.Get("/role/"+roleID+"/identities/", &items); err != nil {
				return graph, err
			}
			graph.Identities = append(graph.Identities, items...)
		}

		if m.Arcs || m.Beats {
			var arcs []map[string]any
			if err := client.Get("/role/"+roleID+"/arcs/", &arcs); err != nil {
				return graph, err
			}
			graph.Arcs = append(graph.Arcs, arcs...)

			if m.Beats {
				for _, arc := range arcs {
					arcID := strField(arc, "id")
					var beats []map[string]any
					if err := client.Get("/arc/"+arcID+"/beats/", &beats); err != nil {
						return graph, err
					}
					graph.Beats = append(graph.Beats, beats...)
				}
			}
		}
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
