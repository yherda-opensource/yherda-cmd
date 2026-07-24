package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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
	Long: `yherda brings your Yherda ideas, persons, and identities into the
terminal — and into whatever else you already work in. Script your workflow,
wire it into an editor or build pipeline, or just work faster without leaving
the keyboard.

Run 'yherda <command> --help' for details on any command. Most commands
operate on an active workspace/idea/person/place/thing, set via the
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
	rootCmd.AddCommand(placeCmd)
	rootCmd.AddCommand(settingCmd)
	rootCmd.AddCommand(thingCmd)
	rootCmd.AddCommand(dispositionCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(modelCmd)
	rootCmd.AddCommand(workspaceListCmd)
	// formatCmd, expressionCmd, and docsCmd are disabled — being replaced. See YOS-80.
	// rootCmd.AddCommand(formatCmd)
	// rootCmd.AddCommand(expressionCmd)
	// rootCmd.AddCommand(docsCmd)
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
	printContextRow(ctx, subjectContextLabel)
}

// printContextRow prints the context summary line, given a function to
// render the Subject field's label (so callers that already fetched the
// Subject once — printContextWithSubject — can reuse that fetch instead of
// calling subjectContextLabel and fetching it again).
func printContextRow(ctx *config.Context, subjectLabel func(string) string) {
	fields := []struct{ label, value string }{
		{"workspace", ctx.Workspace},
		{"idea", ctx.Idea},
		{"person", ctx.Person},
		{"place", ctx.Place},
		{"thing", ctx.Thing},
		{"context", ctx.Context},
		{"state", ctx.State},
		{"goal", ctx.Goal},
		{"subject", subjectLabel(ctx.Subject())},
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

// subjectContextLabel renders the active Subject's full row for the context
// footer — the same fields 'model list' shows (id, name, subject_type,
// has_perspective, has_self) — not just the bare id, since a bare id alone
// doesn't tell the user what they're actually pointed at (see YOS-81).
// Fetched live rather than cached, so a Subject renamed elsewhere doesn't
// show stale data in the footer. Falls back to the bare id if the fetch
// fails, since a secondary lookup failing shouldn't break the command's
// primary output.
func subjectContextLabel(subjectID string) string {
	if subjectID == "" {
		return ""
	}
	creds, err := config.LoadCredentials()
	if err != nil || creds == nil {
		return subjectID
	}
	loadedCtx, err := config.LoadContext()
	if err != nil || loadedCtx.Workspace == "" || loadedCtx.APIServer == "" {
		return subjectID
	}
	client := api.New(loadedCtx.APIServer, creds)
	var subject map[string]any
	if err := client.Get("/subject/"+subjectID+"/", &subject); err != nil {
		return subjectID
	}
	name := strField(subject, "name")
	subjectType := strField(subject, "subject_type")
	if name == "" && subjectType == "" {
		return subjectID
	}
	return fmt.Sprintf("%s %q (%s, has_perspective: %s, has_self: %s)",
		subjectID, name, subjectType, strField(subject, "has_perspective"), strField(subject, "has_self"))
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

// printContextWithSubject prints the two-line footer: the record a command
// just acted on (the Subject, resolved to its name/subject_type) as line 1,
// then the usual context row as line 2. Every model command taking a bare
// Subject id calls this after its own output instead of calling
// printContext() directly, so a wrong-context id (e.g. an Idea id copied
// from 'model list's fallback output) is visible immediately rather than
// silently accepted — not a confirmation prompt, no blocking, just
// visibility, per insight_cli_capability_grant_confirmation_pattern.md.
// Skipped in --json mode, since a JSON caller already knows what it asked
// for. Fetches subjectID once and reuses that row for the context row too
// when it matches the active ctx.Subject, rather than fetching it twice.
func printContextWithSubject(client *api.Client, subjectID string) {
	if noContext || jsonOutput {
		return
	}
	label := subjectID
	var subject map[string]any
	if err := client.Get("/subject/"+subjectID+"/", &subject); err == nil {
		name := strField(subject, "name")
		subjectType := strField(subject, "subject_type")
		fmt.Printf("Subject: #%s %q (%s)\n", subjectID, name, subjectType)
		if name != "" || subjectType != "" {
			label = fmt.Sprintf("%s %q (%s, has_perspective: %s, has_self: %s)",
				subjectID, name, subjectType, strField(subject, "has_perspective"), strField(subject, "has_self"))
		}
	}
	ctx, err := config.LoadContext()
	if err != nil {
		return
	}
	printContextRow(ctx, func(activeSubjectID string) string {
		if activeSubjectID == subjectID {
			return label
		}
		return subjectContextLabel(activeSubjectID)
	})
}

// confirmReader is the source for confirm() prompts. Overridden in tests.
var confirmReader io.Reader = os.Stdin

// confirm prints a y/N prompt and reports whether the user answered yes.
// Only "y" or "yes" (case-insensitive) count as confirmation.
func confirm(prompt string) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(confirmReader)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
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
