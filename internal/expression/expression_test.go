package expression

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// fakeClient answers reads from a path-to-JSON table and records every
// call, so tests can assert exactly what was sent.
type fakeClient struct {
	gets   map[string]string
	posts  map[string]string
	called []string
	bodies []any
}

func (f *fakeClient) Get(path string, out any) error {
	f.called = append(f.called, "GET "+path)
	body, ok := f.gets[path]
	if !ok {
		return fmt.Errorf("API error 404: %s", path)
	}
	return json.Unmarshal([]byte(body), out)
}

func (f *fakeClient) Post(path string, body any, out any) error {
	f.called = append(f.called, "POST "+path)
	f.bodies = append(f.bodies, body)
	resp, ok := f.posts[path]
	if !ok {
		return fmt.Errorf("API error 400: %s", path)
	}
	return json.Unmarshal([]byte(resp), out)
}

func (f *fakeClient) wrote() bool {
	for _, c := range f.called {
		if strings.HasPrefix(c, "PATCH") || strings.HasPrefix(c, "PUT") || strings.HasPrefix(c, "DELETE") {
			return true
		}
	}
	return false
}

const outlineJSON = `[
  {"id": 1, "name": "Package", "type_name": "Package", "number": 1,
   "data_fields": [{"id": 9, "name": "Module Name"}], "layouts": [],
   "children": [
     {"id": 2, "name": null, "type_name": "Version", "number": 1,
      "data_fields": [{"id": 3, "name": "Version Name"}, {"id": 4, "name": "Version Major"}],
      "layouts": [], "children": []},
     {"id": 3, "name": null, "type_name": "Version", "number": 2,
      "data_fields": [], "layouts": [], "children": []}
   ]},
  {"id": 4, "name": "Second", "type_name": "Package", "number": 2,
   "data_fields": [], "layouts": [], "children": []}
]`

func readFake() *fakeClient {
	return &fakeClient{
		gets: map[string]string{
			"/expression/166/segments/": outlineJSON,
			"/expression/166/form/":     `{"id": 7}`,
			"/forminstance/guid-1/?segments=1,2,3,4": `[
			  {"segment_id": 1, "form_data": {"Module Name": "Studio"}, "status": "Active"},
			  {"segment_id": 2, "form_data": {"Version Name": "Studio", "Version Major": "1", "body": "{\"a\": 1}", "Dependencies": "[]"}, "status": "Complete"},
			  {"segment_id": 3, "form_data": {}, "status": "Planned"},
			  {"segment_id": 4, "form_data": {"body": "second"}, "status": "Active"}
			]`,
			"/forminstance/guid-1/?segments=2,3": `[
			  {"segment_id": 2, "form_data": {"Version Name": "Studio"}, "status": "Complete"},
			  {"segment_id": 3, "form_data": {}, "status": "Planned"}
			]`,
		},
		posts: map[string]string{"/form/7/forminstance/": `{"id": "guid-1"}`},
	}
}

func TestList_BuildsTheQuery(t *testing.T) {
	f := &fakeClient{gets: map[string]string{
		"/expression/?idea=5&module=12&type=Package": `[{"id": 1}]`,
	}}
	rows, err := List(f, ListFilter{ModuleID: "12", Type: "Package", IdeaID: "5"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows: %v", rows)
	}
}

func TestList_NoFilterIsTheBarePath(t *testing.T) {
	f := &fakeClient{gets: map[string]string{"/expression/": `[]`}}
	if _, err := List(f, ListFilter{}); err != nil {
		t.Fatal(err)
	}
}

func TestResolveModule_ExactNameOnly(t *testing.T) {
	f := &fakeClient{gets: map[string]string{
		"/module/?name=Module+Builder": `[
		  {"id": 4, "name": "Module Builder Management"},
		  {"id": 9, "name": "Module Builder"}]`,
	}}
	id, err := ResolveModule(f, "Module Builder")
	if err != nil {
		t.Fatal(err)
	}
	if id != "9" {
		t.Errorf("got %q, want the exact-name module 9", id)
	}
}

func TestResolveModule_UnknownIsAnError(t *testing.T) {
	f := &fakeClient{gets: map[string]string{"/module/?name=Nope": `[]`}}
	_, err := ResolveModule(f, "Nope")
	if err == nil || !strings.Contains(err.Error(), `no module named "Nope"`) {
		t.Fatalf("got %v", err)
	}
}

func TestRead_WholeExpression(t *testing.T) {
	f := readFake()
	nodes, err := Read(f, "166", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 || len(nodes[0].Children) != 2 {
		t.Fatalf("shape: %+v", nodes)
	}
	if nodes[0].Children[0].Status != "Complete" {
		t.Errorf("status: %q", nodes[0].Children[0].Status)
	}
	if nodes[0].Children[0].FormData["Version Name"] != "Studio" {
		t.Errorf("content: %v", nodes[0].Children[0].FormData)
	}
	if f.wrote() {
		t.Errorf("read sent a write: %v", f.called)
	}
	// The only POST is the form instance get-or-create.
	posts := 0
	for _, c := range f.called {
		if strings.HasPrefix(c, "POST") {
			posts++
		}
	}
	if posts != 1 {
		t.Errorf("posts: %v", f.called)
	}
}

func TestRead_FromASegmentIsThatSubtree(t *testing.T) {
	// Delete the whole-expression answer so a wrong request fails loudly.
	f := readFake()
	delete(f.gets, "/forminstance/guid-1/?segments=1,2,3,4")
	nodes, err := Read(f, "166", "1")
	if err == nil {
		// segment 1 is the whole first tree: ids 1,2,3
		t.Fatalf("expected the 1,2,3 request to be unanswered, got %+v", nodes)
	}
	f = readFake()
	f.gets["/forminstance/guid-1/?segments=1,2,3"] = `[
	  {"segment_id": 1, "form_data": {}, "status": "Active"},
	  {"segment_id": 2, "form_data": {}, "status": "Active"},
	  {"segment_id": 3, "form_data": {}, "status": "Active"}]`
	nodes, err = Read(f, "166", "1")
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].ID != 1 || len(nodes[0].Children) != 2 {
		t.Errorf("subtree: %+v", nodes)
	}
}

func TestRead_ChildSegmentOnly(t *testing.T) {
	f := readFake()
	nodes, err := Read(f, "166", "2")
	if err != nil {
		// 2 alone has no children: request is segments=2
		f.gets["/forminstance/guid-1/?segments=2"] = `[{"segment_id": 2, "form_data": {"Version Name": "Studio"}, "status": "Complete"}]`
		nodes, err = Read(f, "166", "2")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].ID != 2 {
		t.Errorf("got %+v", nodes)
	}
}

func TestRead_UnknownSegment(t *testing.T) {
	_, err := Read(readFake(), "166", "999")
	if err == nil || !strings.Contains(err.Error(), "segment 999 is not in expression 166") {
		t.Fatalf("got %v", err)
	}
}

func TestRead_NotReadableIsReportedCleanly(t *testing.T) {
	f := readFake()
	delete(f.posts, "/form/7/forminstance/")
	_, err := Read(f, "166", "")
	if err == nil || !strings.Contains(err.Error(), "not readable by you right now") {
		t.Fatalf("got %v", err)
	}
}

func TestRead_BatchesLongIDLists(t *testing.T) {
	// 120 root segments: three batches of at most 50.
	var parts []string
	for i := 1; i <= 120; i++ {
		parts = append(parts, fmt.Sprintf(`{"id": %d, "name": null, "type_name": "T", "number": %d, "data_fields": [], "layouts": [], "children": []}`, i, i))
	}
	f := &fakeClient{
		gets: map[string]string{
			"/expression/5/segments/": "[" + strings.Join(parts, ",") + "]",
			"/expression/5/form/":     `{"id": 2}`,
		},
		posts: map[string]string{"/form/2/forminstance/": `{"id": "g"}`},
	}
	var batches int
	for _, start := range []int{1, 51, 101} {
		end := start + 49
		if end > 120 {
			end = 120
		}
		var ids []string
		for i := start; i <= end; i++ {
			ids = append(ids, fmt.Sprintf("%d", i))
		}
		f.gets["/forminstance/g/?segments="+strings.Join(ids, ",")] = `[]`
		batches++
	}
	if _, err := Read(f, "5", ""); err != nil {
		t.Fatal(err)
	}
	reads := 0
	for _, c := range f.called {
		if strings.HasPrefix(c, "GET /forminstance/") {
			reads++
		}
	}
	if reads != batches {
		t.Errorf("form-instance reads: %d, want %d", reads, batches)
	}
}

func TestRenderDefault_FieldsThenBodyThenOther(t *testing.T) {
	name := "Studio 1.0"
	n := &ReadNode{
		ID: 2, Type: "Version", Name: &name, Status: "Complete",
		FormData: map[string]any{
			"Version Major": "1", "Version Name": "Studio",
			"body": "line one\nline two", "zeta": "z", "alpha": 3.0,
		},
		Node: &Node{DataFields: []DataField{{Name: "Version Name"}, {Name: "Version Major"}}},
	}
	var out bytes.Buffer
	RenderDefault(&out, []*ReadNode{n}, 0)
	want := `[2] Version "Studio 1.0" (Complete)
  Version Name: Studio
  Version Major: 1
  body:
    line one
    line two
  alpha: 3
  zeta: z
`
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestRenderDefault_IndentsChildren(t *testing.T) {
	child := &ReadNode{ID: 2, Type: "Version", FormData: map[string]any{}}
	parent := &ReadNode{ID: 1, Type: "Package", FormData: map[string]any{}, Children: []*ReadNode{child}}
	var out bytes.Buffer
	RenderDefault(&out, []*ReadNode{parent}, 0)
	if !strings.Contains(out.String(), "\n  [2] Version") {
		t.Errorf("child not indented:\n%s", out.String())
	}
}

func TestRenderOutline(t *testing.T) {
	_, roots, err := Outline(readFake(), "166")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	RenderOutline(&out, roots, 0)
	want := "[1] Package 1 \"Package\"\n  [2] Version 1\n  [3] Version 2\n[4] Package 2 \"Second\"\n"
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestJSONOfReadIsRawDataOnly(t *testing.T) {
	nodes, err := Read(readFake(), "166", "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Node") || strings.Contains(string(data), "layouts") {
		t.Errorf("rendering-only fields leaked into JSON: %s", data)
	}
	if !strings.Contains(string(data), `"form_data"`) || !strings.Contains(string(data), `"status"`) {
		t.Errorf("missing data fields: %s", data)
	}
}
