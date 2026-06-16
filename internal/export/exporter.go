package export

import (
	"fmt"
	"sort"
)

// SegmentNode is the intermediate representation shared by all exporters.
type SegmentNode struct {
	TypeName string
	Number   string
	Content  string
	Children []SegmentNode
}

// Exporter serializes a segment tree to a target format.
type Exporter interface {
	Export(title string, roots []SegmentNode, output string) error
	// DefaultExt returns the file extension (without dot) for file-based formats,
	// or "" for stream-based formats like stdout.
	DefaultExt() string
}

var registry = map[string]Exporter{
	"stdout": &StdoutExporter{},
	"scriv":  &ScrivExporter{},
}

// Get returns the exporter for the given format name, or false if unknown.
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

// BuildTree converts the raw JSON segment list (as returned by the API) into
// a []SegmentNode, preserving the nested children structure.
func BuildTree(segs []map[string]any) []SegmentNode {
	nodes := make([]SegmentNode, 0, len(segs))
	for _, seg := range segs {
		node := SegmentNode{
			TypeName: strField(seg, "type_name"),
			Number:   strField(seg, "number"),
			Content:  strField(seg, "segment_data"),
		}
		if children, ok := seg["children"].([]any); ok {
			var childMaps []map[string]any
			for _, c := range children {
				if m, ok := c.(map[string]any); ok {
					childMaps = append(childMaps, m)
				}
			}
			node.Children = BuildTree(childMaps)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func strField(row map[string]any, key string) string {
	if v, ok := row[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
