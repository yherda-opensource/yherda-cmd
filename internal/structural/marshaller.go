package structural

import "fmt"

// Poster is the subset of the API client used by the marshaller.
type Poster interface {
	Post(path string, body any, out any) error
}

// Marshal POSTs an IdeaGraph to the platform API in dependency order:
// identities → places → things → docs.
func Marshal(client Poster, graph IdeaGraph, ideaID string) error {
	// 1. Identities — each needs a person wrapper first.
	for _, identity := range graph.Identities {
		name := strField(identity, "name")

		// Create a person to own this identity.
		var person map[string]any
		if err := client.Post("/storyline/"+ideaID+"/persons/", map[string]string{"name": name}, &person); err != nil {
			return fmt.Errorf("creating person for identity %q: %w", name, err)
		}
		personID := strField(person, "id")

		var created map[string]any
		if err := client.Post("/person/"+personID+"/identities/", map[string]string{"name": name}, &created); err != nil {
			return fmt.Errorf("creating identity %q: %w", name, err)
		}
	}

	// 2. Places.
	for _, place := range graph.Places {
		name := strField(place, "name")
		var created map[string]any
		if err := client.Post("/storyline/"+ideaID+"/places/", map[string]string{"name": name}, &created); err != nil {
			return fmt.Errorf("creating place %q: %w", name, err)
		}
	}

	// 3. Things.
	for _, thing := range graph.Things {
		name := strField(thing, "name")
		var created map[string]any
		if err := client.Post("/storyline/"+ideaID+"/things/", map[string]string{"name": name}, &created); err != nil {
			return fmt.Errorf("creating thing %q: %w", name, err)
		}
	}

	// 4. Docs (idea documents — orphan only; entity-attached docs are future work).
	for _, doc := range graph.Docs {
		title := strField(doc, "title")
		body := strField(doc, "body")
		payload := map[string]string{"title": title}
		if body != "" {
			payload["body"] = body
		}
		var created map[string]any
		if err := client.Post("/storyline/"+ideaID+"/documents/", payload, &created); err != nil {
			return fmt.Errorf("creating document %q: %w", title, err)
		}
	}

	fmt.Println("Import complete.")
	return nil
}
