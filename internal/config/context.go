package config

import (
	"encoding/json"
	"os"
)

type Context struct {
	Workspace    string   `json:"active_workspace,omitempty"`
	APIServer    string   `json:"api_server,omitempty"`
	Idea         string   `json:"active_idea,omitempty"`
	Person       string   `json:"active_person,omitempty"`
	Place        string   `json:"active_place,omitempty"`
	Thing        string   `json:"active_thing,omitempty"`
	Context      string   `json:"active_context,omitempty"`
	State        string   `json:"active_state,omitempty"`
	Goal         string   `json:"active_goal,omitempty"`
	SubjectStack []string `json:"active_subject_stack,omitempty"`
}

// Subject returns the top of the Subject breadcrumb stack, or "" if empty.
func (c *Context) Subject() string {
	if len(c.SubjectStack) == 0 {
		return ""
	}
	return c.SubjectStack[len(c.SubjectStack)-1]
}

// PushSubject pushes id onto the Subject breadcrumb stack, making it active.
func (c *Context) PushSubject(id string) {
	c.SubjectStack = append(c.SubjectStack, id)
}

// PopSubject removes the top of the Subject breadcrumb stack and returns it.
// ok is false if the stack was already empty.
func (c *Context) PopSubject() (id string, ok bool) {
	if len(c.SubjectStack) == 0 {
		return "", false
	}
	last := len(c.SubjectStack) - 1
	id = c.SubjectStack[last]
	c.SubjectStack = c.SubjectStack[:last]
	return id, true
}

// ResetSubject collapses the Subject breadcrumb stack to a single id,
// discarding any existing trail — used when an explicit id override on a
// bare-Subject-id command should start a fresh trail rather than drill
// deeper into the existing one.
func (c *Context) ResetSubject(id string) {
	c.SubjectStack = []string{id}
}

const contextFile = ".yherda"

func LoadContext() (*Context, error) {
	data, err := os.ReadFile(contextFile)
	if os.IsNotExist(err) {
		return &Context{}, nil
	}
	if err != nil {
		return nil, err
	}
	var ctx Context
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}

func SaveContext(ctx *Context) error {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(contextFile, data, 0600)
}
