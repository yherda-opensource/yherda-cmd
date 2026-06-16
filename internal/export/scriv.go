package export

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScrivExporter writes a Scrivener 3 .scriv package from a segment tree.
type ScrivExporter struct{}

func (s *ScrivExporter) DefaultExt() string { return "scriv" }

func (s *ScrivExporter) Export(title string, roots []SegmentNode, output string) error {
	if output == "" {
		output = title + ".scriv"
	}
	if _, err := os.Stat(output); err == nil {
		return fmt.Errorf("output path already exists: %s", output)
	}

	if err := os.MkdirAll(filepath.Join(output, "Files", "Data"), 0755); err != nil {
		return fmt.Errorf("could not create package directory: %w", err)
	}

	var binderItems []scrivItem
	for _, root := range roots {
		item, err := s.buildItem(root, output)
		if err != nil {
			// Clean up partial output on any error.
			os.RemoveAll(output)
			return err
		}
		binderItems = append(binderItems, item)
	}

	if err := s.writeManifest(output, title, binderItems); err != nil {
		os.RemoveAll(output)
		return err
	}

	fmt.Println(output)
	return nil
}

type scrivItem struct {
	UUID     string
	Title    string
	Type     string // "Folder" or "Text"
	Children []scrivItem
}

func (s *ScrivExporter) buildItem(node SegmentNode, pkgPath string) (scrivItem, error) {
	uuid := newUUID()
	itemType := scrivType(node)

	item := scrivItem{
		UUID:  uuid,
		Title: nodeTitle(node),
		Type:  itemType,
	}

	dataDir := filepath.Join(pkgPath, "Files", "Data", uuid)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return scrivItem{}, fmt.Errorf("could not create data dir for %s: %w", uuid, err)
	}

	rtfContent := toRTF(node.Content)
	if err := os.WriteFile(filepath.Join(dataDir, "content.rtf"), []byte(rtfContent), 0644); err != nil {
		return scrivItem{}, fmt.Errorf("could not write content.rtf for %s: %w", uuid, err)
	}

	for _, child := range node.Children {
		childItem, err := s.buildItem(child, pkgPath)
		if err != nil {
			return scrivItem{}, err
		}
		item.Children = append(item.Children, childItem)
	}

	return item, nil
}

func (s *ScrivExporter) writeManifest(pkgPath, title string, items []scrivItem) error {
	type xmlChildren struct {
		Items []any `xml:",any"`
	}

	type xmlBinderItem struct {
		XMLName  xml.Name `xml:"BinderItem"`
		UUID     string   `xml:"UUID,attr"`
		Type     string   `xml:"Type,attr"`
		Title    string   `xml:"Title"`
		Children *xmlChildren
	}

	var buildXML func(items []scrivItem) []any
	buildXML = func(items []scrivItem) []any {
		result := make([]any, 0, len(items))
		for _, item := range items {
			xi := &xmlBinderItem{
				UUID:  item.UUID,
				Type:  item.Type,
				Title: item.Title,
			}
			if len(item.Children) > 0 {
				xi.Children = &xmlChildren{Items: buildXML(item.Children)}
			}
			result = append(result, xi)
		}
		return result
	}

	type scrivxDoc struct {
		XMLName xml.Name `xml:"ScrivenerProject"`
		Version string   `xml:"Version,attr"`
		Binder  struct {
			Items []any `xml:",any"`
		} `xml:"Binder"`
	}

	doc := scrivxDoc{Version: "2.0"}
	doc.Binder.Items = buildXML(items)

	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal manifest: %w", err)
	}

	// Use safe title for filename (replace path separators).
	safeTitle := strings.ReplaceAll(title, "/", "-")
	manifestPath := filepath.Join(pkgPath, safeTitle+".scrivx")
	return os.WriteFile(manifestPath, append([]byte(xml.Header), data...), 0644)
}

func scrivType(node SegmentNode) string {
	if node.Content == "" && len(node.Children) > 0 {
		return "Folder"
	}
	return "Text"
}

func nodeTitle(node SegmentNode) string {
	if node.TypeName != "" && node.Number != "" {
		return node.TypeName + " " + node.Number
	}
	if node.TypeName != "" {
		return node.TypeName
	}
	return "Untitled"
}

func toRTF(content string) string {
	if content == "" {
		return `{\rtf1\ansi {\fonttbl} \pard\par}`
	}
	escaped := strings.ReplaceAll(content, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, "{", `\{`)
	escaped = strings.ReplaceAll(escaped, "}", `\}`)
	return `{\rtf1\ansi {\fonttbl} \pard ` + escaped + `\par}`
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08X-%04X-%04X-%04X-%012X",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
