package structural

import "fmt"

// Poster is the subset of the API client used by the marshaller.
type Poster interface {
	Post(path string, body any, out any) error
}

// Marshal POSTs an IdeaGraph to the platform API in dependency order:
// identities → arcs → beats → places → things → docs.
//
// It maintains an ID map (source _id → platform ID) so child entities
// can reference the correct parent IDs assigned by the API.
func Marshal(client Poster, graph IdeaGraph, ideaID string) error {
	// idMap maps the driver's temporary _id values to platform-assigned IDs.
	idMap := map[string]string{}

	// 1. Identities — each needs a role wrapper first.
	for _, identity := range graph.Identities {
		srcID := strField(identity, "_id")
		name := strField(identity, "name")

		// Create a role to own this identity.
		var role map[string]any
		if err := client.Post("/storyline/"+ideaID+"/roles/", map[string]string{"name": name}, &role); err != nil {
			return fmt.Errorf("creating role for identity %q: %w", name, err)
		}
		roleID := strField(role, "id")

		var created map[string]any
		if err := client.Post("/role/"+roleID+"/identities/", map[string]string{"name": name}, &created); err != nil {
			return fmt.Errorf("creating identity %q: %w", name, err)
		}
		platformID := strField(created, "id")
		if srcID != "" {
			idMap[srcID] = platformID
		}
		// Also map role for arc lookups.
		idMap["role:"+platformID] = roleID
	}

	// 2. Arcs — depend on identity (role).
	for _, arc := range graph.Arcs {
		srcID := strField(arc, "_id")
		name := strField(arc, "name")
		identitySrcID := strField(arc, "_identity_id")

		platformIdentityID := idMap[identitySrcID]
		roleID := idMap["role:"+platformIdentityID]
		if roleID == "" {
			return fmt.Errorf("arc %q references unknown identity %q", name, identitySrcID)
		}

		var created map[string]any
		if err := client.Post("/role/"+roleID+"/arcs/", map[string]string{"name": name}, &created); err != nil {
			return fmt.Errorf("creating arc %q: %w", name, err)
		}
		if srcID != "" {
			idMap[srcID] = strField(created, "id")
		}
	}

	// 3. Beats — depend on arc.
	for _, beat := range graph.Beats {
		name := strField(beat, "name")
		arcSrcID := strField(beat, "_arc_id")
		platformArcID := idMap[arcSrcID]
		if platformArcID == "" {
			return fmt.Errorf("beat %q references unknown arc %q", name, arcSrcID)
		}

		body := map[string]string{"name": name}
		if description := strField(beat, "description"); description != "" {
			body["description"] = description
		}
		var created map[string]any
		if err := client.Post("/arc/"+platformArcID+"/beats/", body, &created); err != nil {
			return fmt.Errorf("creating beat %q: %w", name, err)
		}
	}

	// 4. Places.
	for _, place := range graph.Places {
		name := strField(place, "name")
		var created map[string]any
		if err := client.Post("/storyline/"+ideaID+"/places/", map[string]string{"name": name}, &created); err != nil {
			return fmt.Errorf("creating place %q: %w", name, err)
		}
	}

	// 5. Things.
	for _, thing := range graph.Things {
		name := strField(thing, "name")
		var created map[string]any
		if err := client.Post("/storyline/"+ideaID+"/things/", map[string]string{"name": name}, &created); err != nil {
			return fmt.Errorf("creating thing %q: %w", name, err)
		}
	}

	// 6. Docs (idea documents — orphan only; entity-attached docs are future work).
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
