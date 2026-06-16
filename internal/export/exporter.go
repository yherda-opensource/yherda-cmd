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

// defaultPlugin is the SegmentData layer BuildTree reads content from. Each
// segment can carry one SegmentData row per plugin (writer today; an
// editor/co-writer plugin later) — this is hardcoded until a second plugin
// exists to design a --plugin flag against.
const defaultPlugin = "writer"

// BuildTree converts the raw JSON segment list (as returned by the API) into
// a []SegmentNode, preserving the nested children structure. Content is read
// from the defaultPlugin's entry in each segment's "data" array.
func BuildTree(segs []map[string]any) []SegmentNode {
	return buildTreeForPlugin(segs, defaultPlugin)
}

func buildTreeForPlugin(segs []map[string]any, plugin string) []SegmentNode {
	nodes := make([]SegmentNode, 0, len(segs))
	for _, seg := range segs {
		node := SegmentNode{
			TypeName: strField(seg, "type_name"),
			Number:   strField(seg, "number"),
			Content:  contentForPlugin(seg, plugin),
		}
		if children, ok := seg["children"].([]any); ok {
			var childMaps []map[string]any
			for _, c := range children {
				if m, ok := c.(map[string]any); ok {
					childMaps = append(childMaps, m)
				}
			}
			node.Children = buildTreeForPlugin(childMaps, plugin)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// contentForPlugin returns the segment_data content for the given plugin's
// entry in a segment's "data" array, or "" if no such entry exists.
func contentForPlugin(seg map[string]any, plugin string) string {
	data, ok := seg["data"].([]any)
	if !ok {
		return ""
	}
	for _, entry := range data {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if strField(m, "plugin") == plugin {
			return strField(m, "segment_data")
		}
	}
	return ""
}

func strField(row map[string]any, key string) string {
	if v, ok := row[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
