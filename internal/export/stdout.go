package export

import "fmt"

// StdoutExporter prints segment content depth-first to stdout.
// Structural nodes with no content are skipped (no decorations).
type StdoutExporter struct{}

func (s *StdoutExporter) DefaultExt() string { return "" }

func (s *StdoutExporter) Export(title string, roots []SegmentNode, output string) error {
	var walk func(nodes []SegmentNode)
	walk = func(nodes []SegmentNode) {
		for _, n := range nodes {
			if n.Content != "" {
				fmt.Println(n.Content)
			}
			walk(n.Children)
		}
	}
	walk(roots)
	return nil
}
