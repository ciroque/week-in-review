# week-in-review

Generate an append-only chronological project journal from Git repository metadata.

`week-in-review` is intentionally deterministic. It does not use an LLM, a Git hosting API, or filesystem timestamps. Git history is the source of truth.

## What it does

For a completed Monday-through-Sunday reporting period, the tool:

- reads commits from a Git repository;
- records commit and author timestamps;
- records commit SHA, author, and subject;
- calculates simple activity statistics;
- appends one Markdown section to a journal file;
- refuses to overwrite a week that is already present.

The journal is designed to live in the repository it describes.

## Install

### Go

```bash
go install github.com/ciroque/week-in-review/cmd/week-in-review@latest
```

### Docker

```bash
docker run --rm \
  -v "$PWD:/repo" \
  ghcr.io/ciroque/week-in-review:latest \
  --repo /repo
```

## Usage

```text
week-in-review [flags]

Flags:
  --init                 Initialize a new journal and exit\n  --repo string          Git repository to inspect (default ".")
  --output string        Journal path, relative to --repo (default "WEEK-IN-REVIEW.md")
  --timezone string      IANA timezone used for reporting boundaries (default "America/Los_Angeles")
  --week-ending string   Sunday ending the reporting week, YYYY-MM-DD; defaults to the most recently completed Sunday
  --ref string           Git ref to inspect (default "HEAD")
```

Example:

```bash
week-in-review \
  --repo . \
  --output WEEK-IN-REVIEW.md \
  --timezone America/Los_Angeles \
  --week-ending 2026-10-11
```

## Design constraints

- No network is required.
- No GitHub/GitLab API is required.
- No LLM is used.
- No filesystem timestamps are used.
- Existing week sections are treated as immutable.
- Duplicate reporting periods fail with a non-zero exit code.
- Calendar boundaries are evaluated in the configured IANA timezone.
- Binary-file numstat entries are ignored for line totals.

## CI/CD examples

Ready-to-adapt examples are included for both major hosting platforms:

- `examples/github-actions.yml` — scheduled GitHub Actions workflow using `GITHUB_TOKEN` to commit and push the journal.
- `examples/gitlab-ci.yml` — scheduled GitLab pipeline using a `WEEK_IN_REVIEW_PUSH_TOKEN` CI/CD variable with `write_repository` permission.

Both examples fetch full Git history, run `week-in-review`, and commit `WEEK-IN-REVIEW.md` only when it changes.

For provenance-sensitive use, pin a released container image tag rather than `:latest`.

## License

MIT.
