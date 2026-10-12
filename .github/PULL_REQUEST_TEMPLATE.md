## Summary

<!-- What does this PR do, in one or two sentences? -->

## Related issue

<!-- Link the issue(s) this closes, e.g. "Closes #12". "None" if N/A. -->

## Changes

<!-- Bullet list of what changed and why. Keep it factual. -->

- 

## Testing done

- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Manual run against a local specs repo (paste the command + key output)

## Safety checklist

- [ ] Parse → Render round-trip stays byte-identical for canonical spec files
- [ ] Mutating commands support `--dry-run` and never write on dry-run
- [ ] `evidence` stays fail-closed (non-zero exit never checks an AC)
- [ ] New behavior defaults to *off* (fail-closed) unless intentionally opt-in
- [ ] README/CHANGELOG updated if commands, flags, or file format changed
