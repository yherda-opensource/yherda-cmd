package structural

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ObsidianExporter writes an IdeaGraph as a vault of .md files.
type ObsidianExporter struct{}

func (o *ObsidianExporter) DefaultExt() string { return "" }

func (o *ObsidianExporter) Export(graph IdeaGraph, output string) error {
	if output == "" {
		output = "obsidian-export"
	}

	// Index docs by entity_type+entity_id for O(1) lookup.
	docIndex := buildDocIndex(graph.Docs)

	type entityGroup struct {
		subdir        string
		entities      []map[string]any
		frontMatterFn func(e map[string]any) map[string]string
	}

	groups := []entityGroup{
		{
			subdir:   "identities",
			entities: graph.Identities,
			frontMatterFn: func(e map[string]any) map[string]string {
				return map[string]string{
					"id":      strField(e, "id"),
					"created": strField(e, "created"),
				}
			},
		},
		{
			subdir:   "places",
			entities: graph.Places,
			frontMatterFn: func(e map[string]any) map[string]string {
				return map[string]string{
					"id":      strField(e, "id"),
					"created": strField(e, "created"),
				}
			},
		},
		{
			subdir:   "things",
			entities: graph.Things,
			frontMatterFn: func(e map[string]any) map[string]string {
				return map[string]string{
					"id":      strField(e, "id"),
					"created": strField(e, "created"),
				}
			},
		},
	}

	for _, g := range groups {
		dir := filepath.Join(output, g.subdir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		seen := map[string]int{}
		for _, entity := range g.entities {
			name := nameForEntity(entity)
			filename := uniqueFilename(dir, name, seen)
			body := docBodyForEntity(docIndex, g.subdir, strField(entity, "id"))
			fm := g.frontMatterFn(entity)
			if err := writeMarkdown(filepath.Join(dir, filename), fm, body); err != nil {
				return err
			}
		}
	}

	// Orphan docs — not attached to any entity.
	orphanDir := filepath.Join(output, "docs")
	if err := os.MkdirAll(orphanDir, 0755); err != nil {
		return err
	}
	seen := map[string]int{}
	for _, doc := range graph.Docs {
		if strField(doc, "entity_type") != "" {
			continue
		}
		title := strField(doc, "title")
		if title == "" {
			title = strField(doc, "id")
		}
		filename := uniqueFilename(orphanDir, title, seen)
		fm := map[string]string{
			"id":      strField(doc, "id"),
			"created": strField(doc, "created"),
		}
		body := strField(doc, "body")
		if err := writeMarkdown(filepath.Join(orphanDir, filename), fm, body); err != nil {
			return err
		}
	}

	fmt.Printf("Exported to %s\n", output)
	return nil
}

// buildDocIndex indexes docs by "entityType:entityID" for fast lookup.
func buildDocIndex(docs []map[string]any) map[string]map[string]any {
	idx := map[string]map[string]any{}
	for _, doc := range docs {
		et := strField(doc, "entity_type")
		eid := strField(doc, "entity_id")
		if et != "" && eid != "" {
			idx[et+":"+eid] = doc
		}
	}
	return idx
}

// docBodyForEntity returns the body from the IdeaDocument attached to an entity, or "".
func docBodyForEntity(idx map[string]map[string]any, entityType, entityID string) string {
	if doc, ok := idx[entityType+":"+entityID]; ok {
		return strField(doc, "body")
	}
	return ""
}

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

// sanitizeFilename converts a title to a lowercase kebab-case filename (no extension).
func sanitizeFilename(title string) string {
	lower := strings.ToLower(title)
	slug := nonAlphanumRe.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "untitled"
	}
	return slug
}

// uniqueFilename returns a collision-free filename (with .md extension),
// appending -2, -3, etc. as needed. seen tracks base slugs used in a directory.
func uniqueFilename(dir, title string, seen map[string]int) string {
	base := sanitizeFilename(title)
	seen[base]++
	if seen[base] == 1 {
		return base + ".md"
	}
	return fmt.Sprintf("%s-%d.md", base, seen[base])
}

// nameForEntity returns the display name of an entity, preferring "name" then "title".
func nameForEntity(e map[string]any) string {
	if v := strField(e, "name"); v != "" {
		return v
	}
	if v := strField(e, "title"); v != "" {
		return v
	}
	return strField(e, "id")
}

// writeMarkdown writes a YAML front-matter block followed by body content.
func writeMarkdown(path string, frontMatter map[string]string, body string) error {
	var sb strings.Builder
	sb.WriteString("---\n")
	keys := []string{"id", "created"}
	for _, k := range keys {
		if v, ok := frontMatter[k]; ok && v != "" {
			fmt.Fprintf(&sb, "%s: %s\n", k, v)
		}
	}
	sb.WriteString("---\n")
	if body != "" {
		sb.WriteString("\n")
		sb.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			sb.WriteString("\n")
		}
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
