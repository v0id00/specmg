package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// skillMarkdown is the SKILL.md specmg prints for AI agents. Keep command
// signatures in sync with the real CLI surface (verified by tests).
const skillMarkdown = `---
name: specmg
description: "Use when managing project specs via the specmg CLI (create, status, requirements/ACs, lint, verify, report)."
---

# specmg — spec management

NEVER edit files under specs/ by hand. Always use the specmg CLI from the
project root (it finds specs/ by walking up, like git).

## Spec file format (specs/NNNN-slug.md)

    ---
    id: "0001"
    slug: retry-budget
    title: Retry budget
    status: draft
    owner: alice
    created: 2026-10-12
    updated: 2026-10-12
    tags: [backend]
    ---

Sections after frontmatter: Purpose (free prose), Requirements
(- **REQ-N**: ...), Links (- **L-N**: rel -> NNNN-slug), Acceptance Criteria
(- [ ] **AC-N**: ...), Open Questions (- **Q-N** [blocking]: ...),
Decision Log (markdown table).
The CLI re-renders files canonically on every write — hand formatting is lost.

Status lifecycle: draft -> in_review -> approved -> implemented -> verified
(also superseded, rejected). Link rels: implements, relates, blocks,
depends_on, supersedes. Ref = id (0001), slug, or filename.

## Commands

- specmg init — create specs/ if missing
- specmg new "<title>" [--owner x] [--tag t] — create a Draft spec
- specmg get <ref> [--json] — full spec (human = rendered markdown)
- specmg list [--status s]... [--owner o] [--tag t] [--json] — filtered overview
- specmg set <ref> [--status s] [--title t] [--owner o] [--tag t] [--slug s]
- specmg req add <ref> "<text>" | rm <ref> REQ-N | list <ref>
- specmg ac add <ref> "<text>" | check|uncheck <ref> AC-N | list <ref>
- specmg link add <ref> --rel <rel> --to <target> | rm <ref> L-N | list <ref>
- specmg q add <ref> "<text>" [--blocking] | resolve <ref> Q-N | list <ref>
- specmg log <ref> "<decision>" [--rationale r] — append a Decision Log row
- specmg supersede <ref> --by <other> | specmg rm <ref>
- specmg lint [--strict] [--json] — consistency checks, exit 1 on errors
- specmg stats [--json] | specmg tree [<ref>] | specmg search <text>
- specmg export [--csv] | specmg index [--write] | specmg import
- specmg verify <ref> — run backticked commands in all open ACs; check only on exit 0
- specmg evidence <ref> --ac AC-N --cmd "<shell>" — run once, store output, check on exit 0
- specmg report [<ref>...] [--format html|pdf] [--out f] [--title t]
- specmg skill — print this file; specmg version; specmg completion <shell>

Global flags: --json, --dry-run, --plain (no icons), --root <dir>, -v.
Exit codes: 0 ok, 1 error, 2 usage.

## Rules

1. Every mutating command supports --dry-run; never skip it for bulk changes.
2. Only the USER moves a spec to approved — never self-approve.
3. An AC must be checkable by one command or one observable behavior; prefer
   backtick-wrapped commands as AC text, then use verify/evidence.
4. Blocking open questions must be resolved (q resolve) before approval;
   specmg lint enforces this.
5. Record decisions with log, never by editing the Decision Log table.
6. Cite other specs by id (link add) instead of copying their contract.
`

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Print the specmg SKILL.md for AI agents",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprint(cmd.OutOrStdout(), skillMarkdown)
	},
}

func init() { rootCmd.AddCommand(skillCmd) }
