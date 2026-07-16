package config

import (
	"encoding/json"
	"os"
)

type Context struct {
	Workspace string `json:"active_workspace,omitempty"`
	APIServer string `json:"api_server,omitempty"`
	Idea      string `json:"active_idea,omitempty"`
	Person    string `json:"active_person,omitempty"`
	Place     string `json:"active_place,omitempty"`
	Thing     string `json:"active_thing,omitempty"`
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
