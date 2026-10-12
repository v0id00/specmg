# Contributing to specmg

Thanks for helping out! specmg is a small, focused tool — it manages spec files
that projects treat as their source of truth. That shapes every rule below:
correctness and file safety matter more than speed.

## Before you start

1. **Search first.** Check [issues](https://github.com/v0id00/specmg/issues) for
   an existing report or request before opening a new one.
2. **Ask before big changes.** If a feature changes the CLI surface or the spec
   file format, open an issue and discuss the design before writing code. Small
   fixes and docs are welcome directly.
3. **One concern per PR.** A PR that fixes a bug and refactors a package is two
   PRs.

## Reporting bugs

Open an issue with:

- **specmg version** (`specmg version`) and OS
- **The exact command** you ran
- **The spec file** involved (paste it — specs are not secret by default; say
  so if yours are)
- **Expected vs. actual output**, including the error

## Feature requests

Open an issue describing:

- The **problem** you're solving (not just the feature)
- How it fits specmg's existing commands — specmg prefers one obvious way per
  operation, `--json` everywhere
- Alternatives you considered

## Development setup

Requirements: Go 1.24+, `make`.

```bash
git clone https://github.com/v0id00/specmg.git
cd specmg
make build    # binary in ./specmg
make test     # go test ./...
make lint     # gofmt -l . && go vet ./...
```

Project layout:

```
cmd/specmg/        entrypoint
internal/spec/     spec model, markdown parse/render (canonical format)
internal/store/    filesystem, atomic writes, id/slug resolution
internal/cli/      cobra commands
internal/output/   JSON/table output helpers
```

## Code style

- `gofmt` clean, standard library first, existing patterns followed
- Comments in English, matching the codebase
- Add or update tests for behavior changes — `internal/cli/cli_test.go` shows
  the command-test pattern (temp repo + `run()` helper)

## Commit conventions

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add --stream live output
fix: respect --dry-run on evidence
docs: clarify canonical render behavior
test: cover lint duplicate detection
chore: bump go.mod deps
```

Imperative mood, lowercase scope when useful (`feat(cli): ...`).

## Pull requests

1. Branch from `main`: `git checkout -b fix/lint-dup`
2. Make the change, run `make test` and `make lint` until green
3. Push and open a PR against `main` — the CI workflow runs lint + tests
4. Fill in the PR template; keep the description factual

## Safety checklist (important)

specmg rewrites spec files it owns. Before submitting:

- [ ] `Parse` → `Render` round-trip stays byte-identical for canonical files
      (`internal/spec` tests)
- [ ] All mutating commands still support `--dry-run` and never write on dry-run
- [ ] `evidence` stays fail-closed: non-zero exit must never check an AC
- [ ] Defaults stay fail-closed: new risky behavior must default to *off*
- [ ] `--json` added/updated for any new read command, `--help` text accurate
- [ ] README and CHANGELOG updated if commands, flags, or file format changed

## Code of conduct

All contributors are expected to follow the
[Code of Conduct](CODE_OF_CONDUCT.md). Report unacceptable behavior through the
channels listed there.
