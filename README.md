# specmg

Manage specs, requirements and acceptance criteria from the CLI — no manual
editing of `specs/` files.

```
go build -o specmg ./cmd/specmg
```

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

Statuses: `draft | in_review | approved | implemented | verified | superseded | rejected`.
Link rels: `implements | relates | blocks | depends_on | supersedes`.
Files are owned by the CLI: every write re-renders the file in canonical form.

## Commands

```
specmg init                                create specs/ if missing
specmg new "<title>" [--owner x] [--tag t] create a Draft spec
specmg get <id|slug> [--json]              full spec (human = raw markdown)
specmg list [--status s]... [--owner o] [--tag t] [--json]
specmg set <id|slug> [--status s] [--title t] [--owner o] [--tag t] [--slug s]
specmg req add <ref> "<text>" | rm <ref> REQ-N | list <ref>
specmg ac  add <ref> "<text>" | check|uncheck <ref> AC-N | list <ref>
specmg link add <ref> --rel <rel> --to <target> | rm <ref> L-N | list <ref>
specmg q   add <ref> "<text>" [--blocking] | resolve <ref> Q-N | list <ref>
specmg supersede <ref> --by <other>
specmg rm <ref>
specmg lint [--strict] [--json]            consistency checks; exit 1 on errors
specmg stats [--json]                      counts by status, open ACs/questions
specmg tree [<ref>]                        link graph (ASCII)
specmg evidence <ref> --ac AC-N --cmd "<shell>"   run cmd, store output, check AC on exit 0
specmg skill                               print SKILL.md for AI agents
specmg version
```

`<ref>` is an id (`0001`), slug (`retry-budget`) or full filename.
Global flags: `--json`, `--dry-run` (all writes), `--root <dir>`, `-v`.

Exit codes: `0` ok, `1` error (incl. lint problems), `2` usage error.

## evidence

Runs `--cmd` via `sh -c` (deliberate: full shell commands), stores stdout/stderr
and exit code in `specs/evidence/<id>-AC-N-<ts>.md`, and checks the AC only on
exit 0 (fail-closed). `--dry-run` never executes.

## Development

```
make build   # go build -o specmg ./cmd/specmg
make test    # go test ./...
make lint    # gofmt -l . && go vet ./...
```
