package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func credentialsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(home, ".yherdacmd")
	if err := os.MkdirAll(d, 0700); err != nil {
		return "", err
	}
	return d, nil
}

func LoadCredentials() (*Credentials, error) {
	d, err := credentialsDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(d, "credentials.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

func SaveCredentials(creds *Credentials) error {
	d, err := credentialsDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, "credentials.json"), data, 0600)
}

func DeleteCredentials() error {
	d, err := credentialsDir()
	if err != nil {
		return err
	}
	path := filepath.Join(d, "credentials.json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
