# Security Policy

## Supported versions

Only the latest release is supported. Security fixes land on `main` and are
shipped in the next release.

## Reporting a vulnerability

**Do not open a public issue for security problems.**

Report privately via the GitHub Security Advisory:

https://github.com/v0id00/specmg/security/advisories/new

Please include:

- Affected version(s)
- Impact — what an attacker can do, and under what conditions
- Reproduction steps (spec file, command)
- Suggested fix, if you have one

You'll get an acknowledgement within a few days and a timeline for the fix and
release.

## Security notes for users

- **`evidence` runs shell commands**: `specmg evidence --cmd "<shell>"` executes
  your command via `sh -c` by design. Only run commands you trust; never let an
  untrusted source (spec text from the internet, a CI log, a prompt) choose the
  `--cmd` value. `--dry-run` prints the command without executing it.
- **Spec files are code-adjacent**: AC text often contains commands
  (backtick-wrapped). Treat spec files from untrusted repos like you would a
  script — read before running `evidence` against them.
- **No network, no telemetry**: specmg reads and writes files under `specs/`
  only and executes nothing except an explicit `evidence --cmd`.
