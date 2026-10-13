# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-10-13

### Fixed

- `import` now understands real sdd files: numbered section headings
  (`## 7. Acceptance Criteria`), items wrapped over several lines, `AC1 —` /
  `OQ1 —` ids and resolved (`~~OQ1 — …~~ **Resolved …**`) questions. Original
  item numbers and checkbox state are preserved, the `Last updated` line fills
  `created`/`updated`, and the header block above the first heading survives as
  `## Metadata` instead of becoming an unnamed section.
- Documents in `specs/` that are not specs (`README.md`, `INDEX.md`, anything
  not named `NNNN-slug.md`) are ignored by every command instead of failing the
  run with `missing YAML frontmatter`.

### Added

- `verify <ref>` — run backticked commands in all open ACs, check on exit 0 (fail-closed), PASS/FAIL report
- `log <ref> "<decision>" [--rationale]` — append a dated row to the Decision Log
- `search <text>` — search titles, purposes and item texts, `--json` supported
- `export [--csv]` — flat rows of every spec and item
- `index [--write]` — print or write `specs/INDEX.md`
- `import` — convert legacy sdd-format spec files to specmg format in place
- `completion` — shell completion (cobra built-in)
- Pretty output mode: status badges (🟡🔵🟢🟣✅⚫🔴) and item icons (🔹 ☐/☑ ❓❗)
  on human output; `--plain` disables, `--json` never contains icons
- `report [<ref>...]` — standalone HTML or PDF report for one or more specs
  (`--format`, `--out`, `--title`); PDF embeds a system UTF-8 font when found

## [0.1.0] - 2026-10-12

First release.

### Added

- Spec CRUD: `init`, `new`, `get`, `set` (status/title/owner/tags/slug rename), `rm`
- Item CRUD: `req add|rm|list`, `ac add|check|uncheck|list`, `link add|rm|list`, `q add|resolve|list`
- Status lifecycle: `draft → in_review → approved → implemented → verified` (+ `superseded`, `rejected`)
- `supersede <ref> --by <other>` — retire a spec with a Decision Log entry
- `list` with `--status`/`--owner`/`--tag` filters; `--json` on every read command
- `lint [--strict]` — dangling links, blocking questions on approved specs, duplicate item numbers; exit 1 on errors (CI-friendly)
- `stats` — counts by status, open ACs and questions
- `tree [<ref>]` — ASCII link graph with backlinks
- `evidence <ref> --ac AC-N --cmd "<shell>"` — runs the command, stores output under `specs/evidence/`, checks the AC only on exit 0 (fail-closed)
- `skill` — prints a SKILL.md for AI agents
- Global flags: `--json`, `--dry-run`, `--root`, `-v`
- Markdown + YAML frontmatter storage format (`specs/NNNN-<slug>.md`), canonical re-render on every write
- GitHub Actions release workflow: tag push builds linux/windows amd64 archives with checksums
