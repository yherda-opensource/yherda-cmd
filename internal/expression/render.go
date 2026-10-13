package expression

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// bodyKey is the key the default editor stores free text under for any
// segment type, whether or not the type declares fields.
const bodyKey = "body"

// RenderDefault writes a segment and its descendants as text: a heading,
// then the type's declared fields in order, then the body, then any stored
// keys that are not declared fields (so nothing stored is hidden). It is
// the rendering used for a segment whose type defines no layout for the
// channel.
func RenderDefault(w io.Writer, nodes []*ReadNode, depth int) {
	for _, n := range nodes {
		renderNode(w, n, depth)
		RenderDefault(w, n.Children, depth+1)
	}
}

func renderNode(w io.Writer, n *ReadNode, depth int) {
	pad := strings.Repeat("  ", depth)
	label := n.Type
	if n.Name != nil && *n.Name != "" {
		label += " " + fmt.Sprintf("%q", *n.Name)
	}
	status := n.Status
	if status != "" {
		status = " (" + status + ")"
	}
	fmt.Fprintf(w, "%s[%d] %s%s\n", pad, n.ID, label, status)

	shown := map[string]bool{}
	var declared []string
	if n.Node != nil {
		for _, f := range n.Node.DataFields {
			declared = append(declared, f.Name)
		}
	}
	for _, key := range declared {
		shown[key] = true
		if v, ok := n.FormData[key]; ok {
			writeField(w, pad+"  ", key, v)
		}
	}
	if v, ok := n.FormData[bodyKey]; ok && !shown[bodyKey] {
		shown[bodyKey] = true
		writeField(w, pad+"  ", bodyKey, v)
	}
	var others []string
	for key := range n.FormData {
		if !shown[key] {
			others = append(others, key)
		}
	}
	sort.Strings(others)
	for _, key := range others {
		writeField(w, pad+"  ", key, n.FormData[key])
	}
}

func writeField(w io.Writer, pad, key string, value any) {
	text := valueText(value)
	if !strings.Contains(text, "\n") {
		fmt.Fprintf(w, "%s%s: %s\n", pad, key, text)
		return
	}
	fmt.Fprintf(w, "%s%s:\n", pad, key)
	for _, line := range strings.Split(text, "\n") {
		fmt.Fprintf(w, "%s  %s\n", pad, line)
	}
}

func valueText(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		data, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(data)
	}
}

// RenderOutline writes the tree: id, type, name and number, indented by
// depth.
func RenderOutline(w io.Writer, nodes []*Node, depth int) {
	for _, n := range nodes {
		pad := strings.Repeat("  ", depth)
		label := n.TypeName
		if n.Number != nil {
			label += fmt.Sprintf(" %d", *n.Number)
		}
		if n.Name != nil && *n.Name != "" {
			label += " " + fmt.Sprintf("%q", *n.Name)
		}
		fmt.Fprintf(w, "%s[%d] %s\n", pad, n.ID, label)
		RenderOutline(w, n.Children, depth+1)
	}
}
