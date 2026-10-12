# specmg

[![ci](https://github.com/v0id00/specmg/actions/workflows/ci.yml/badge.svg)](https://github.com/v0id00/specmg/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/v0id00/specmg)](https://github.com/v0id00/specmg/releases/latest)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

CLI to manage **specs, requirements and acceptance criteria** — create, status,
link, lint and verify them without opening a spec file in an editor. Built for
both humans and AI agents.

```console
$ specmg new "Retry budget" --owner alice --tag backend
created specs/0001-retry-budget.md
$ specmg req add 0001 "The client SHALL retry at most 3 times."
added REQ-1 to 0001-retry-budget
$ specmg list
ID    STATUS  TITLE          OWNER  REQS  OPEN-AC  OPEN-Q
0001  draft   Retry budget   alice  1     0        0
```

## Why

Specs usually live in markdown files that drift out of date because editing
them by hand is friction. `specmg` makes the CLI the interface: every change is
validated (statuses, link targets, item ids), rendered in a canonical format,
and stays git-friendly markdown you can still read and diff.

## Install

### Prebuilt binaries

Download from the [latest release](https://github.com/v0id00/specmg/releases/latest)
(`specmg_<version>_linux_amd64.tar.gz` or `specmg_<version>_windows_amd64.zip`),
verify the checksum against `checksums-<os>.txt`, extract, and put `specmg` on
your `PATH`.

```console
$ tar xzf specmg_0.1.0_linux_amd64.tar.gz
$ ./specmg_0.1.0_linux_amd64/specmg version
specmg 0.1.0
```

### From source

```console
$ go install github.com/v0id00/specmg/cmd/specmg@latest
```

or clone and `make build` (binary lands in the repo root).

## Quick start

```console
$ specmg init                                # creates specs/ (run inside your repo)
$ specmg new "Retry budget" --owner alice --tag backend
created specs/0001-retry-budget.md
$ specmg req add 0001 "The client SHALL retry at most 3 times."
$ specmg ac  add 0001 "`specmg list` exits 0 on an empty specs dir."
$ specmg q   add 0001 "Should backoff be exponential?" --blocking
$ specmg set 0001 --status in_review
$ specmg q   resolve 0001 Q-1
$ specmg set 0001 --status approved
$ specmg evidence 0001 --ac AC-1 --cmd "specmg list"
AC-1 verified (evidence: evidence/0001-AC-1-1791807070.md)
$ specmg lint
lint: ok
```

`specmg` finds `specs/` by walking up from the current directory (like git), or
takes an explicit `--root <dir>`.

## Spec file format

One spec per file: `specs/NNNN-<slug>.md` (4-digit zero-padded id).

```markdown
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

## Purpose

Prose, preserved verbatim.

## Requirements

- **REQ-1**: The client SHALL retry failed calls at most 3 times.

## Links

- **L-1**: implements -> 0002-other-slug

## Acceptance Criteria

- [ ] **AC-1**: `specmg list` exits 0.
- [x] **AC-2**: Already verified.

## Open Questions

- **Q-1** [blocking]: Should backoff be exponential?
- **Q-2** [resolved]: What is the max retry count?

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-10-12 | Max 3 retries | Simplicity |
```

- **Statuses**: `draft | in_review | approved | implemented | verified | superseded | rejected`
- **Link rels**: `implements | relates | blocks | depends_on | supersedes`
- `<ref>` arguments accept an id (`0001`), a slug (`retry-budget`) or the full
  filename (`0001-retry-budget.md`).
- Files are owned by the CLI: every write re-renders the file in canonical
  form. `Purpose`, `Decision Log` and unknown sections are preserved verbatim;
  hand-edited item formatting may be normalized.
- Item ids (`REQ-N`, `AC-N`, `Q-N`, `L-N`) are assigned automatically
  (`max + 1` per type).

## Commands

| Command | Description |
|---------|-------------|
| `specmg init` | Create `specs/` if missing |
| `specmg new "<title>" [--owner x] [--tag t]` | Create a Draft spec |
| `specmg get <ref> [--json]` | Full spec (human output = raw markdown) |
| `specmg list [--status s]... [--owner o] [--tag t] [--json]` | Filterable overview |
| `specmg set <ref> [--status s] [--title t] [--owner o] [--tag t] [--slug s]` | Update fields; `--slug` renames the file |
| `specmg req add\|rm\|list` | Requirements CRUD |
| `specmg ac add\|check\|uncheck\|list` | Acceptance criteria CRUD |
| `specmg link add --rel <rel> --to <target>\|rm\|list` | Cross-spec links (target is validated) |
| `specmg q add [--blocking]\|resolve\|list` | Open questions |
| `specmg supersede <ref> --by <other>` | Retire a spec; adds a Decision Log row |
| `specmg rm <ref>` | Delete a spec file |
| `specmg lint [--strict] [--json]` | Consistency checks; exit 1 on errors |
| `specmg stats [--json]` | Counts by status, open ACs/questions |
| `specmg tree [<ref>]` | ASCII link graph (backlinks with a ref) |
| `specmg evidence <ref> --ac AC-N --cmd "<shell>"` | Run command, store output, check AC on exit 0 |
| `specmg verify <ref> [--all]` | Run backticked commands in all open ACs; check on exit 0, report pass/fail |
| `specmg log <ref> "<decision>" [--rationale r]` | Append a row to the Decision Log |
| `specmg search <text> [--case]` | Search titles, purposes and item texts (`--json` supported) |
| `specmg export [--csv]` | Flat rows of every spec + item (JSON default) |
| `specmg index [--write]` | Print or write `specs/INDEX.md` |
| `specmg import [--dir d]` | Convert legacy sdd-format specs to specmg format (in place) |
| `specmg completion bash\|zsh\|fish\|powershell` | Shell completion script (cobra built-in) |
| `specmg skill` | Print a SKILL.md for AI agents |
| `specmg version` | Print version |

**Global flags**: `--json` (machine-readable output on every command),
`--dry-run` (all writes print what they would do instead), `--plain` (disable
icons/status badges), `--root <dir>`, `-v`.

**Pretty output**: human output is icon-decorated by default — status badges
(🟡 draft, 🔵 in_review, 🟢 approved, 🟣 implemented, ✅ verified, ⚫ superseded,
🔴 rejected), requirement 🔹, checkboxes ☐/☑, questions ❓ open / ❗ blocking /
✅ resolved, verify ✅ PASS / ❌ FAIL. `--plain` turns all of it off; `--json`
never contains icons.

**Exit codes**: `0` ok · `1` error (including lint problems) · `2` usage error.

## Lint

`specmg lint` is CI-friendly (`--json` output, exit 1 on errors):

- approved spec with an unresolved **blocking** question
- dangling link target (points at a spec that no longer exists)
- duplicate item numbers within a spec
- `superseded` spec missing `superseded_by` (warning; error with `--strict`)

```console
$ specmg lint
specs/0001-retry-budget.md: [error] approved spec has unresolved blocking Q-1
error: lint: 1 problem(s)
$ echo $?
1
```

## evidence

`specmg evidence <ref> --ac AC-N --cmd "<shell>"` runs the command via `sh -c`
(deliberate: full shell commands like `go test ./... && make lint` are the
point), stores stdout/stderr and the exit code in
`specs/evidence/<id>-AC-N-<timestamp>.md`, and checks the AC **only on exit 0**
(fail-closed). `--dry-run` prints what it would do without executing anything.

## verify

`specmg verify <ref>` scans every unchecked AC for a backtick-wrapped command
(`\`true\` exits 0`), runs each via `sh -c`, checks the AC on exit 0 and reports
`PASS`/`FAIL` per criterion plus a summary. Non-zero exits fail the command
(exit 1) without checking that AC. ACs without a command are listed with
`--all`. Like `evidence`, `--dry-run` never executes.

```console
$ specmg verify 0001
PASS AC-1: specmg list
FAIL AC-2: make lint (exit 1)
verified 2 AC(s): 1 passed, 1 failed
```

## Decision Log

`specmg log <ref> "<decision>" [--rationale "..."]` appends a dated row to the
spec's Decision Log table (creating the table header if missing).

## Using with AI agents

```console
$ specmg skill > .claude/skills/specmg/SKILL.md   # or your agent's skill dir
```

The generated skill tells the agent to never edit `specs/` by hand, always use
the CLI, never self-approve `Draft → Approved`, and only check ACs with real
evidence.

## Development

```console
make build   # go build -o specmg ./cmd/specmg
make test    # go test ./...
make lint    # gofmt -l . && go vet ./...
```

Layout: `internal/spec` (model, markdown parse/render), `internal/store`
(filesystem, atomic writes, id/slug resolution), `internal/cli` (cobra
commands), `internal/output` (JSON/table).

Releases: push a `v*` tag — CI runs tests and publishes linux/windows amd64
archives with per-OS sha256 checksums.

## License

[MIT](LICENSE)
