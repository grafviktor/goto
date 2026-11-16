---
name: generate-release-notes
description: Writes RELEASE_NOTES_vX.Y.Z.md for one application version by copying that version's CHANGELOG.md section and grouping its bullets into Improvements, Fixes, and Technical. Use when the user invokes generate-release-notes or asks for release notes for a version such as v1.7.0 or 1.7.0.
disable-model-invocation: true
---

# Generate Release Notes

Write one release-notes file for a single application version.

`generate-release-notes $APP_VERSION`

Examples: `generate-release-notes v1.7.0` or `generate-release-notes 1.7.0`.

If the version is missing, ask for it and stop. Do not invent a version and do not take it from the top of `CHANGELOG.md`.

Read [references/example.md](references/example.md) before writing and match its layout. The versions, issues, and notes in that file are mock data.

## Version

Accept `vMAJOR.MINOR.PATCH` or `MAJOR.MINOR.PATCH`. Reject anything else and ask again.

- `$APP_VERSION` is the argument, with a leading `v` (`1.7.0` becomes `v1.7.0`).
- `X.Y.Z` is the major, minor, and patch of `$APP_VERSION` (`v1.7.0` uses `1.7.0`).

Write the file at the repository root as `RELEASE_NOTES_vX.Y.Z.md`. Example: `RELEASE_NOTES_v1.7.0.md`. If that file already exists, replace it.

## Source

Copy the release from `CHANGELOG.md`. Find the section whose header links to `v$APP_VERSION`. If that section is missing, say so and stop. Do not use an older release.

Include every bullet in that section: task bullets, `_Notes_` bullets, and `_Breaking changes_` bullets. Do not drop one. Do not add an issue that is not in that section.

Read the milestone before classifying or writing the summary:

https://github.com/grafviktor/goto/issues?q=milestone%3AvX.Y.Z

```bash
gh issue list --repo grafviktor/goto --milestone "vX.Y.Z" --state all --limit 200 --json number,title,body,labels
```

When `gh` is unavailable:

```bash
curl -fsSL --get "https://api.github.com/search/issues" \
  --data-urlencode "q=repo:grafviktor/goto milestone:vX.Y.Z is:issue" \
  --data-urlencode "per_page=100"
```

Match a changelog details link to the issue with the same number. Use the issue title, body, and labels only to choose a category and to write the summary and breaking-change instructions. Keep each copied bullet's wording and details link.

## Categories

Put every task bullet under `## What's changed ##`, in this order: `### Improvements ###`, `### Fixes ###`, `### Technical ###`. Omit a category that has no bullets.

- **Fixes**: the bullet starts with `Fix`, or the issue is labeled bug, or its title is prefixed with `BUG-` or `FIX-`, or the change corrects broken behavior.
- **Technical**: the change has no user-visible behavior (build, dependencies, CI, refactor, tests) and still appears in the changelog.
- **Improvements**: every other task bullet.

A task bullet with no issue number stays in the category its wording supports. When the category is unclear, use Improvements.

Keep `[[details](https://github.com/grafviktor/goto/issues/$ISSUE_NUMBER)].` exact, including the trailing period. A bullet with no issue number has no details link.

## Summary, notes, and breaking changes

Start the file with one short paragraph that states what this release changes for the user. Write it from the changelog bullets. Do not invent a feature.

After the categories, add `## Notes ##` only when the changelog section has `_Notes_` bullets. Copy those bullets unchanged. Leave one blank line before the heading.

Add `## Breaking changes ##` only when the changelog section has `_Breaking changes_` bullets. Copy the facts from those bullets, then write the user instruction and a fenced shell command when the changelog names a command. Follow [references/example.md](references/example.md). Omit the section when there is nothing to undo.

End the file with `<hr>`, followed by full change log.

Edit only `RELEASE_NOTES_vX.Y.Z.md`. Never commit.
