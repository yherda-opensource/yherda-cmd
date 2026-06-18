package structural

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ScrivImporter parses a Scrivener 3 .scriv package into an IdeaGraph.
type ScrivImporter struct{}

func (s *ScrivImporter) Import(source string) (IdeaGraph, error) {
	scrivxPath, err := findScrivx(source)
	if err != nil {
		return IdeaGraph{}, err
	}

	binder, err := parseScrivx(scrivxPath)
	if err != nil {
		return IdeaGraph{}, fmt.Errorf("parsing %s: %w", scrivxPath, err)
	}

	dataDir := filepath.Join(source, "Files", "Data")
	return mapBinderToGraph(binder, dataDir), nil
}

// scrivx XML structures.
// Structure: <ScrivenerProject><Binder><BinderItem ...><Children><BinderItem ...>
// The Binder element contains a single root BinderItem; the top-level working
// folders are its Children. parseScrivx returns those children directly.

type scrivxFile struct {
	BinderRoot binderItem `xml:"Binder>BinderItem"`
}

type binderItem struct {
	UUID     string       `xml:"UUID,attr"`
	Type     string       `xml:"Type,attr"`
	Title    string       `xml:"Title"`
	Children []binderItem `xml:"Children>BinderItem"`
}

func findScrivx(source string) (string, error) {
	entries, err := os.ReadDir(source)
	if err != nil {
		return "", fmt.Errorf("opening %s: %w", source, err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".scrivx") {
			return filepath.Join(source, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no .scrivx file found in %s", source)
}

func parseScrivx(path string) ([]binderItem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f scrivxFile
	if err := xml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.BinderRoot.Children, nil
}

// mapBinderToGraph applies mapping heuristics to the top-level binder folders.
func mapBinderToGraph(items []binderItem, dataDir string) IdeaGraph {
	graph := IdeaGraph{PluginData: map[string]any{}}

	for _, item := range items {
		titleLower := strings.ToLower(item.Title)

		switch {
		case isCharacterFolder(titleLower):
			for _, child := range item.Children {
				srcID := child.UUID
				name := child.Title
				if name == "" {
					name = srcID
				}
				identity := map[string]any{
					"_id":  srcID,
					"name": name,
				}
				graph.Identities = append(graph.Identities, identity)
			}

		case isManuscriptFolder(titleLower):
			// Sub-folders are arcs; scenes inside are beats.
			// If there are no sub-folders, treat direct children as beats under a single arc.
			arcFolders := foldersIn(item.Children)
			if len(arcFolders) == 0 {
				// Single implicit arc named after the manuscript folder.
				arcSrcID := item.UUID + "-arc"
				arc := map[string]any{
					"_id":          arcSrcID,
					"_identity_id": "", // marshaller will skip if empty
					"name":         item.Title,
				}
				graph.Arcs = append(graph.Arcs, arc)
				for _, scene := range item.Children {
					beat := map[string]any{
						"_arc_id":     arcSrcID,
						"name":        scene.Title,
						"description": readContent(dataDir, scene.UUID),
					}
					graph.Beats = append(graph.Beats, beat)
				}
			} else {
				for _, arcFolder := range arcFolders {
					arcSrcID := arcFolder.UUID
					arc := map[string]any{
						"_id":          arcSrcID,
						"_identity_id": "", // no identity mapping from manuscript folders
						"name":         arcFolder.Title,
					}
					graph.Arcs = append(graph.Arcs, arc)
					for _, scene := range arcFolder.Children {
						beat := map[string]any{
							"_arc_id":     arcSrcID,
							"name":        scene.Title,
							"description": readContent(dataDir, scene.UUID),
						}
						graph.Beats = append(graph.Beats, beat)
					}
				}
			}

		case isResearchFolder(titleLower):
			for _, child := range allLeaves(item.Children) {
				title := child.Title
				if title == "" {
					title = child.UUID
				}
				doc := map[string]any{
					"title": title,
					"body":  readContent(dataDir, child.UUID),
				}
				graph.Docs = append(graph.Docs, doc)
			}

		default:
			// Unmapped — preserve as idea documents, flagged for dry-run reporting.
			for _, child := range allLeaves(item.Children) {
				title := child.Title
				if title == "" {
					title = child.UUID
				}
				doc := map[string]any{
					"title":     title,
					"body":      readContent(dataDir, child.UUID),
					"_unmapped": "true",
				}
				graph.Docs = append(graph.Docs, doc)
			}
			// Also include the folder itself if it has no children.
			if len(item.Children) == 0 {
				doc := map[string]any{
					"title":     item.Title,
					"body":      readContent(dataDir, item.UUID),
					"_unmapped": "true",
				}
				graph.Docs = append(graph.Docs, doc)
			}
		}
	}

	return graph
}

func isCharacterFolder(title string) bool {
	return title == "characters" || title == "character sheets" || title == "cast"
}

func isManuscriptFolder(title string) bool {
	return title == "manuscript" || title == "draft" || title == "story"
}

func isResearchFolder(title string) bool {
	return title == "research" || title == "notes" || title == "references"
}

func foldersIn(items []binderItem) []binderItem {
	var folders []binderItem
	for _, item := range items {
		if item.Type == "Folder" || len(item.Children) > 0 {
			folders = append(folders, item)
		}
	}
	return folders
}

// allLeaves returns all leaf documents (no children) recursively.
func allLeaves(items []binderItem) []binderItem {
	var leaves []binderItem
	for _, item := range items {
		if len(item.Children) == 0 {
			leaves = append(leaves, item)
		} else {
			leaves = append(leaves, allLeaves(item.Children)...)
		}
	}
	return leaves
}

// readContent reads the plain-text body of a binder item from Files/Data/{UUID}/.
// It tries content.rtf first, then content.xml, returning "" on any failure.
func readContent(dataDir, uuid string) string {
	itemDir := filepath.Join(dataDir, uuid)

	rtfPath := filepath.Join(itemDir, "content.rtf")
	if data, err := os.ReadFile(rtfPath); err == nil {
		return stripRTF(string(data))
	}

	xmlPath := filepath.Join(itemDir, "content.xml")
	if data, err := os.ReadFile(xmlPath); err == nil {
		return stripXMLTags(string(data))
	}

	return ""
}

var rtfControlRe = regexp.MustCompile(`\\[a-z*]+[-\d]*\s?|[{}]`)
var rtfBinRe = regexp.MustCompile(`\\\*\\[a-z]+[^}]*}`)

// stripRTF extracts plain text from an RTF string.
// This is intentionally minimal — formatting fidelity is out of scope.
func stripRTF(rtf string) string {
	// Remove binary data blocks first.
	text := rtfBinRe.ReplaceAllString(rtf, "")
	// Remove RTF control words and group delimiters.
	text = rtfControlRe.ReplaceAllString(text, "")
	// Unescape RTF special characters.
	text = strings.ReplaceAll(text, "\\{", "{")
	text = strings.ReplaceAll(text, "\\}", "}")
	text = strings.ReplaceAll(text, "\\\\", "\\")
	return strings.TrimSpace(text)
}

var xmlTagRe = regexp.MustCompile(`<[^>]+>`)

func stripXMLTags(s string) string {
	return strings.TrimSpace(xmlTagRe.ReplaceAllString(s, ""))
}
