package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var jsonOutput bool
var noContext bool

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
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output raw JSON")
	rootCmd.PersistentFlags().BoolVar(&noContext, "no-context", false, "Suppress context footer")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(ideasCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(personCmd)
	rootCmd.AddCommand(identityCmd)
	rootCmd.AddCommand(arcCmd)
	rootCmd.AddCommand(beatCmd)
	rootCmd.AddCommand(placeCmd)
	rootCmd.AddCommand(settingCmd)
	rootCmd.AddCommand(thingCmd)
	rootCmd.AddCommand(dispositionCmd)
	rootCmd.AddCommand(workspaceListCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(expressionCmd)
	rootCmd.AddCommand(docsCmd)
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fatalf("failed to serialize output: %v", err)
	}
	fmt.Println(string(data))
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
}

func strField(row map[string]any, key string) string {
	if v, ok := row[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func printContext() {
	if noContext || jsonOutput {
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}
	fields := []struct{ label, value string }{
		{"workspace", cfg.ActiveWorkspace},
		{"idea", cfg.ActiveIdea},
		{"person", cfg.ActivePerson},
		{"arc", cfg.ActiveArc},
		{"place", cfg.ActivePlace},
		{"thing", cfg.ActiveThing},
	}
	var parts []string
	for _, f := range fields {
		if f.value != "" {
			parts = append(parts, f.label+": "+f.value)
		}
	}
	if len(parts) > 0 {
		fmt.Println()
		fmt.Println("context: " + joinStrings(parts, " | "))
	}
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func mustPublicClient() *api.Client {
	creds, err := config.LoadCredentials()
	if err != nil {
		fatalf("failed to load credentials: %v", err)
	}
	if creds == nil {
		fatalf("not logged in — run 'yherda login' first")
	}
	return api.NewPublic(creds)
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
	if cfg.APIServer == "" {
		fatalf("no API server configured — run 'yherda workspace <slug>' to reconfigure")
	}
	return api.New(cfg.APIServer, creds)
}
