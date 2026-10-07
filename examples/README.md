# CI/CD Examples

This directory contains ready-to-adapt examples for running `week-in-review` automatically from GitHub Actions and GitLab CI/CD.

## GitHub Actions

See [`github-actions.yml`](github-actions.yml).

Copy it into your repository as:

```text
.github/workflows/week-in-review.yml
```

The workflow:

- runs on a Monday schedule and can also be started manually;
- checks out the repository with full Git history (`fetch-depth: 0`);
- runs the published `week-in-review` container;
- commits `WEEK-IN-REVIEW.md` only when it changes;
- pushes the update using the built-in `GITHUB_TOKEN` with `contents: write`.

Adjust the cron schedule and reporting timezone as needed.

## GitLab CI/CD

See [`gitlab-ci.yml`](gitlab-ci.yml).

Either incorporate the job into your existing `.gitlab-ci.yml` or include it from your pipeline configuration.

The job:

- runs for scheduled pipelines and can also be started manually from the GitLab UI;
- sets `GIT_DEPTH: "0"` so the complete repository history is available;
- runs the published `week-in-review` container;
- commits `WEEK-IN-REVIEW.md` only when it changes;
- pushes the update back to the default branch.

Create a masked CI/CD variable named `WEEK_IN_REVIEW_PUSH_TOKEN` containing a project or group access token with `write_repository` permission.

## Initial bootstrap

Before enabling either scheduled job, initialize an existing repository once:

```bash
week-in-review --repo . --init
```

This creates `WEEK-IN-REVIEW.md` from the repository's complete reachable commit history, grouped into Monday-through-Sunday reporting periods. Weeks without commits are omitted.

Commit the generated journal normally:

```bash
git add WEEK-IN-REVIEW.md
git commit -m "Initialize week in review"
git push
```

Subsequent CI/CD runs use the normal incremental mode and append the most recently completed week.

## Container versions

The examples use:

```text
ghcr.io/ciroque/week-in-review:latest
```

For reproducible or provenance-sensitive workflows, replace `:latest` with a specific released version tag.
