package cmd

// YOS-88: the `expression` group for finding and reading expressions
// (list, use, outline, read). It lives in its own file with its own
// identifiers: cmd/expression.go is the legacy command YOS-80 is replacing,
// currently disabled and left untouched; it uses the same `expression`
// name, so YOS-80 reconciles the two when it lands.

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
	"github.com/yherda-opensource/yherda-cmd/internal/expression"
)

var exprGroupCmd = &cobra.Command{
	Use:   "expression",
	Short: "Find and read expressions",
	Long: `An expression is a piece of written work in an idea: a screenplay, a
novel, a module's package. These commands find the expressions you can see
and read them the way the web app's todo screen does.

You see the expressions in ideas you own and the ones you hold an occupation
on. Reading never changes anything.`,
}

var (
	exprModuleName string
	exprType       string
	exprIdeaID     string
	exprFlagID     string
	exprSegmentID  string
)

// activeExpression resolves the expression a command works on: the
// --expression flag, else the active expression from the context. It does
// not guess when neither is set.
func activeExpression(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	ctx, err := config.LoadContext()
	if err != nil {
		return "", err
	}
	if ctx.Expression == "" {
		return "", fmt.Errorf("no active expression — run 'yherda expression use <id>' or pass --expression <id>")
	}
	return ctx.Expression, nil
}

func yesNo(v any) string {
	if b, ok := v.(bool); ok && b {
		return "yes"
	}
	return "no"
}

var exprListCmd = &cobra.Command{
	Use:   "list",
	Short: "List expressions, by module or type",
	Long: `Lists the expressions you can see in the workspace. Filter by the module
their format belongs to (--module, the module's exact name, quoted when it has
spaces), by their format (--type), or by idea (--idea).`,
	Example: `  yherda expression list
  yherda expression list --module "Module Builder"
  yherda expression list --type Package --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		filter := expression.ListFilter{Type: exprType, IdeaID: exprIdeaID}
		if exprModuleName != "" {
			id, err := expression.ResolveModule(client, exprModuleName)
			if err != nil {
				return err
			}
			filter.ModuleID = id
		}
		rows, err := expression.List(client, filter)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(rows)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tMODULE\tIDEA\tOWNER\tCREATED")
		for _, row := range rows {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				strField(row, "id"), strField(row, "template_name"),
				strField(row, "format_name"), strField(row, "module_name"),
				strField(row, "idea"), yesNo(row["is_owner"]),
				strField(row, "created"))
		}
		w.Flush()
		printContext()
		return nil
	},
}

var exprUseCmd = &cobra.Command{
	Use:     "use <id>",
	Short:   "Set the active expression",
	Long:    "Sets the active expression, stored in the .yherda context file, after checking that you can read it.",
	Example: `  yherda expression use 166`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("expression id %q is not a number", args[0])
		}
		if _, err := expression.Get(mustClient(), args[0]); err != nil {
			return fmt.Errorf("cannot read expression %s: %w", args[0], err)
		}
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Expression = args[0]
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active expression set to %s\n", args[0])
		return nil
	},
}

var exprOutlineCmd = &cobra.Command{
	Use:   "outline",
	Short: "Show an expression's outline",
	Long: `Prints the expression's segment tree: id, type, number and name, indented
by depth. Use the ids with 'expression read --segment'. Uses the active
expression unless --expression is passed.`,
	Example: `  yherda expression outline
  yherda expression outline --expression 166 --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := activeExpression(exprFlagID)
		if err != nil {
			return err
		}
		raw, roots, err := expression.Outline(mustClient(), id)
		if err != nil {
			return err
		}
		if jsonOutput {
			os.Stdout.Write(raw)
			fmt.Println()
			return nil
		}
		expression.RenderOutline(os.Stdout, roots, 0)
		printContext()
		return nil
	},
}

var exprReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Read an expression, from any segment",
	Long: `Reads an expression top to bottom, or, with --segment, that segment and all
its children. Each segment shows its type's fields, then its body. With --json
the raw stored content and status are printed instead. Uses the active
expression unless --expression is passed.`,
	Example: `  yherda expression read
  yherda expression read --expression 166 --segment 542
  yherda expression read --segment 542 --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := activeExpression(exprFlagID)
		if err != nil {
			return err
		}
		nodes, err := expression.Read(mustClient(), id, exprSegmentID)
		if err != nil {
			return err
		}
		if jsonOutput {
			data, err := json.MarshalIndent(nodes, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}
		expression.RenderDefault(os.Stdout, nodes, 0)
		printContext()
		return nil
	},
}

func init() {
	exprListCmd.Flags().StringVar(&exprModuleName, "module", "", "Module name, exactly (quote it when it has spaces)")
	exprListCmd.Flags().StringVar(&exprType, "type", "", "Format name, e.g. Package")
	exprListCmd.Flags().StringVar(&exprIdeaID, "idea", "", "Idea ID")
	exprOutlineCmd.Flags().StringVar(&exprFlagID, "expression", "", "Expression ID (overrides active context)")
	exprReadCmd.Flags().StringVar(&exprFlagID, "expression", "", "Expression ID (overrides active context)")
	exprReadCmd.Flags().StringVar(&exprSegmentID, "segment", "", "Segment ID: read that segment and its children")
	exprGroupCmd.AddCommand(exprListCmd, exprUseCmd, exprOutlineCmd, exprReadCmd)
}
