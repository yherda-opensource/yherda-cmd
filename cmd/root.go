package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/auth"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var jsonOutput bool
var noContext bool

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "yherda",
	Short: "Command line interface for Yherda",
	Long: `yherda brings your Yherda ideas, identities, arcs, and beats into the
terminal — and into whatever else you already work in. Script your workflow,
wire it into an editor or build pipeline, or just work faster without leaving
the keyboard.

Run 'yherda <command> --help' for details on any command. Most commands
operate on an active workspace/idea/person/arc/place/thing, set via the
'use' subcommands and stored in a .yherda file in the current directory, so
you rarely need to pass IDs by hand once you've set context.

Every command supports --json for raw, pipeable output — handy for scripts,
and for AI agents that want to drive Yherda directly.`,
	Version: version,
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
	ctx, err := config.LoadContext()
	if err != nil {
		return
	}
	fields := []struct{ label, value string }{
		{"workspace", ctx.Workspace},
		{"idea", ctx.Idea},
		{"person", ctx.Person},
		{"arc", ctx.Arc},
		{"place", ctx.Place},
		{"thing", ctx.Thing},
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

// useParent persists a parent id to the working directory context when it was
// supplied explicitly on the command line. Pass a setter that writes the id
// into the context struct.
func useParent(set func(*config.Context, string), id string) {
	ctx, err := config.LoadContext()
	if err != nil {
		return
	}
	set(ctx, id)
	_ = config.SaveContext(ctx)
}

func makeRefreshFunc(creds *config.Credentials) func() (*config.Credentials, error) {
	return func() (*config.Credentials, error) {
		newCreds, err := auth.RefreshTokens(creds.RefreshToken)
		if err != nil {
			return nil, err
		}
		if err := config.SaveCredentials(newCreds); err != nil {
			return nil, err
		}
		*creds = *newCreds
		return newCreds, nil
	}
}

func mustPublicClient() *api.Client {
	creds, err := config.LoadCredentials()
	if err != nil {
		fatalf("failed to load credentials: %v", err)
	}
	if creds == nil {
		fatalf("not logged in — run 'yherda login' first")
	}
	c := api.NewPublic(creds)
	c.RefreshFunc = makeRefreshFunc(creds)
	return c
}

func mustClient() *api.Client {
	creds, err := config.LoadCredentials()
	if err != nil {
		fatalf("failed to load credentials: %v", err)
	}
	if creds == nil {
		fatalf("not logged in — run 'yherda login' first")
	}
	ctx, err := config.LoadContext()
	if err != nil {
		fatalf("failed to load context: %v", err)
	}
	if ctx.Workspace == "" {
		fatalf("no active workspace — run 'yherda workspace <slug>'")
	}
	if ctx.APIServer == "" {
		fatalf("no API server configured — run 'yherda workspace <slug>' to reconfigure")
	}
	c := api.New(ctx.APIServer, creds)
	c.RefreshFunc = makeRefreshFunc(creds)
	return c
}
