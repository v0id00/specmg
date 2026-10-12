package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Print the specmg SKILL.md for AI agents",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprint(cmd.OutOrStdout(), skillMarkdown)
	},
}

// skillMarkdown is the SKILL.md content for AI agent consumption.
const skillMarkdown = "---\n" +
	"name: specmg\n" +
	"description: \"Use when managing project specs via the specmg CLI (create, update status, add requirements/ACs, list).\"\n" +
	"---\n\n" +
	"## Rules\n" +
	"- NEVER edit specs/ by hand — use the CLI\n" +
	"- Status lifecycle: draft → in_review → approved → implemented → verified\n" +
	"- Only the USER approves specifications\n" +
	"- ACs must be command-or-observable\n" +
	"- Blocking questions must be resolved before approval\n" +
	"- Evidence only recorded with real command output\n\n" +
	"## Commands\n\n" +
	"- **new** — Create a new spec file from stdin\n" +
	"- **get <ref>** — Print a spec's rendered content\n" +
	"- **list** — List all specs with IDs, slugs, and statuses\n" +
	"- **set <ref> --status <s>** — Change a spec's status\n" +
	"- **req <ref> --text \"...\"** — Add a requirement\n" +
	"- **ac <ref> --text \"...\"** — Add an acceptance criterion\n" +
	"- **link <ref> --rel <rel> --target <id>** — Add a link\n" +
	"- **q <ref> --text \"...\" [--blocking]** — Add an open question\n" +
	"- **supersede <ref> --by <other-ref>** — Mark a spec as superseded\n" +
	"- **rm <ref>** — Remove a spec file\n" +
	"- **lint [--strict] [--json]** — Lint all specs for issues\n" +
	"- **stats [--json]** — Aggregate statistics across specs\n" +
	"- **tree [<ref>]** — Show links as a tree; with arg show backlinks too\n" +
	"- **evidence <ref> --ac <AC-N> --cmd <c>** — Run shell command, record evidence\n" +
	"- **verify <ref>** — Run backticked commands in all open ACs; only exit-0 commands get checked\n" +
	"- **log <ref> \"<decision>\" [--rationale r]** — Append a dated Decision Log row\n" +
	"- **search <text>** — Search titles, purposes and item texts\n" +
	"- **export [--csv]** — Flat rows of every spec and item\n" +
	"- **index [--write]** — Print or write specs/INDEX.md\n" +
	"- **import** — Convert legacy sdd-format specs to specmg format\n" +
	"- **report [<ref>...] --format html|pdf --out <file>** — Render specs as a shareable HTML/PDF report\n" +
	"- **init** — Create the specs/ directory if missing\n" +
	"- **version** — Print version\n\n" +
	"All mutating commands support dry-run mode (--dry-run). Most commands support --json for machine-readable output. Human output uses status badges and item icons; pass --plain for icon-free output (and --json never contains icons).\n"

func init() {
	rootCmd.AddCommand(skillCmd)
}
