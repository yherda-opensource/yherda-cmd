// Package expression reads expressions through the platform's generic,
// owner/occupation-scoped reads: list them (optionally by module or type),
// fetch an outline, and read segment content the way the web todo screen
// does (through the expression's form instance). Everything here is
// read-only; nothing knows what a Package, a Layout or any other format is.
package expression

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Client is the part of api.Client this package needs; *api.Client
// satisfies it, and tests supply a fake.
type Client interface {
	Get(path string, out any) error
	Post(path string, body any, out any) error
}

// ListFilter narrows the expression list. Empty fields do not filter.
type ListFilter struct {
	ModuleID string // module id, as GET /expression/?module=
	Type     string // format name, as GET /expression/?type=
	IdeaID   string // idea id, as GET /expression/?idea=
}

// List returns the expressions the caller can see, as the raw rows the
// platform returned (so --json can print them untouched).
func List(c Client, f ListFilter) ([]map[string]any, error) {
	q := url.Values{}
	if f.ModuleID != "" {
		q.Set("module", f.ModuleID)
	}
	if f.Type != "" {
		q.Set("type", f.Type)
	}
	if f.IdeaID != "" {
		q.Set("idea", f.IdeaID)
	}
	path := "/expression/"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var rows []map[string]any
	if err := c.Get(path, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ResolveModule returns the id of the module with exactly this name. The
// platform answers GET /module/?name= with the installed active version
// first and the caller's own in-progress module when none is installed;
// the first exact match wins. No name normalization.
func ResolveModule(c Client, name string) (string, error) {
	var rows []map[string]any
	path := "/module/?name=" + url.QueryEscape(name)
	if err := c.Get(path, &rows); err != nil {
		return "", err
	}
	for _, row := range rows {
		if n, _ := row["name"].(string); n == name {
			if id, ok := numberString(row["id"]); ok {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("no module named %q is visible to you", name)
}

// Get fetches one expression row (also how `use` checks that the caller
// can read it).
func Get(c Client, id string) (map[string]any, error) {
	var row map[string]any
	if err := c.Get("/expression/"+id+"/", &row); err != nil {
		return nil, err
	}
	return row, nil
}

// DataField is one of a segment type's declared fields.
type DataField struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Panel is one panel of a layout.
type Panel struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Number  int    `json:"number"`
	Content string `json:"content"`
}

// Layout is a layout a segment type defines for a channel and mode.
type Layout struct {
	ID              int     `json:"id"`
	Channel         string  `json:"channel"`
	Mode            string  `json:"mode"`
	Panels          []Panel `json:"panels"`
	SummaryTemplate string  `json:"summary_template"`
}

// Node is one segment of an expression's outline.
type Node struct {
	ID         int         `json:"id"`
	Name       *string     `json:"name"`
	TypeName   string      `json:"type_name"`
	Number     *int        `json:"number"`
	DataFields []DataField `json:"data_fields"`
	Layouts    []Layout    `json:"layouts"`
	Children   []*Node     `json:"children"`
}

// Outline fetches an expression's whole segment tree. raw is the
// platform's JSON untouched, for --json.
func Outline(c Client, expressionID string) (raw json.RawMessage, roots []*Node, err error) {
	if err = c.Get("/expression/"+expressionID+"/segments/", &raw); err != nil {
		return nil, nil, err
	}
	if err = json.Unmarshal(raw, &roots); err != nil {
		return nil, nil, fmt.Errorf("unexpected outline response: %w", err)
	}
	return raw, roots, nil
}

// Find returns the node with this id anywhere in the forest, or nil.
func Find(roots []*Node, id int) *Node {
	for _, n := range roots {
		if n.ID == id {
			return n
		}
		if found := Find(n.Children, id); found != nil {
			return found
		}
	}
	return nil
}

// ReadNode is a segment with its stored content and status. Node carries
// the type's fields and layouts for rendering and is not part of the JSON.
type ReadNode struct {
	ID       int            `json:"id"`
	Type     string         `json:"type_name"`
	Name     *string        `json:"name"`
	Number   *int           `json:"number"`
	Status   string         `json:"status"`
	FormData map[string]any `json:"form_data"`
	Children []*ReadNode    `json:"children"`

	Node *Node `json:"-"`
}

type segmentContent struct {
	SegmentID int            `json:"segment_id"`
	FormData  map[string]any `json:"form_data"`
	Status    string         `json:"status"`
}

type form struct {
	ID int `json:"id"`
}

type formInstance struct {
	ID string `json:"id"`
}

// batchSize keeps the segment ids of one form-instance read short enough
// for a URL.
const batchSize = 50

// Read reads an expression's content: the whole expression when
// segmentID is empty, otherwise that segment and all its children. It
// resolves the expression's form and its form instance (get-or-create, as
// the web app does) and sends no write beyond that.
func Read(c Client, expressionID, segmentID string) ([]*ReadNode, error) {
	_, roots, err := Outline(c, expressionID)
	if err != nil {
		return nil, err
	}
	start := roots
	if segmentID != "" {
		id, convErr := strconv.Atoi(segmentID)
		if convErr != nil {
			return nil, fmt.Errorf("segment id %q is not a number", segmentID)
		}
		node := Find(roots, id)
		if node == nil {
			return nil, fmt.Errorf("segment %s is not in expression %s", segmentID, expressionID)
		}
		start = []*Node{node}
	}

	var ids []int
	var collect func(nodes []*Node)
	collect = func(nodes []*Node) {
		for _, n := range nodes {
			ids = append(ids, n.ID)
			collect(n.Children)
		}
	}
	collect(start)

	content := map[int]segmentContent{}
	if len(ids) > 0 {
		var f form
		if err := c.Get("/expression/"+expressionID+"/form/", &f); err != nil {
			return nil, notReadable(expressionID, err)
		}
		var fi formInstance
		body := map[string]any{"expression_id": mustInt(expressionID)}
		if err := c.Post(fmt.Sprintf("/form/%d/forminstance/", f.ID), body, &fi); err != nil {
			return nil, notReadable(expressionID, err)
		}
		for i := 0; i < len(ids); i += batchSize {
			end := i + batchSize
			if end > len(ids) {
				end = len(ids)
			}
			parts := make([]string, 0, end-i)
			for _, id := range ids[i:end] {
				parts = append(parts, strconv.Itoa(id))
			}
			var batch []segmentContent
			path := "/forminstance/" + fi.ID + "/?segments=" + strings.Join(parts, ",")
			if err := c.Get(path, &batch); err != nil {
				return nil, notReadable(expressionID, err)
			}
			for _, item := range batch {
				content[item.SegmentID] = item
			}
		}
	}
	return attach(start, content), nil
}

func attach(nodes []*Node, content map[int]segmentContent) []*ReadNode {
	out := make([]*ReadNode, 0, len(nodes))
	for _, n := range nodes {
		item := content[n.ID]
		data := item.FormData
		if data == nil {
			data = map[string]any{}
		}
		out = append(out, &ReadNode{
			ID: n.ID, Type: n.TypeName, Name: n.Name, Number: n.Number,
			Status: item.Status, FormData: data,
			Children: attach(n.Children, content), Node: n,
		})
	}
	return out
}

func notReadable(expressionID string, err error) error {
	return fmt.Errorf("expression %s is not readable by you right now: %w", expressionID, err)
}

func mustInt(s string) any {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return s
}

func numberString(v any) (string, bool) {
	switch n := v.(type) {
	case float64:
		return strconv.FormatInt(int64(n), 10), true
	case string:
		return n, n != ""
	}
	return "", false
}
