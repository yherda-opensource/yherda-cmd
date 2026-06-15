package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var pretty bool

var rootCmd = &cobra.Command{
	Use:   "yherda",
	Short: "Command line interface for Yherda",
	Long:  "A fully functioning, tenant-aware CLI for Yherda. Pipeable JSON output for agent and script use.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "Pretty-print JSON output")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(ideasCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(identitiesCmd)
	rootCmd.AddCommand(arcsCmd)
	rootCmd.AddCommand(beatsCmd)
}

func printJSON(v any) {
	var data []byte
	var err error
	if pretty {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		data, err = json.Marshal(v)
	}
	if err != nil {
		fatalf("failed to serialize output: %v", err)
	}
	fmt.Println(string(data))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func mustClient() *api.Client {
	creds, err := config.LoadCredentials()
	if err != nil {
		fatalf("failed to load credentials: %v", err)
	}
	if creds == nil {
		fatalf("not logged in — run 'yherda login' first")
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		fatalf("failed to load config: %v", err)
	}
	if cfg.ActiveWorkspace == "" {
		fatalf("no active workspace — run 'yherda workspace <slug>'")
	}
	return api.New(cfg.ActiveWorkspace, creds)
}
