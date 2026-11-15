---
name: update-changelog
description: Prepends a release section to CHANGELOG.md from two version arguments, previous then current (v1.6.1 v1.7.0, or 1.6.1 1.7.0). Use when the user invokes update-changelog, asks to update CHANGELOG.md for a release, or names the previous and current application versions.
disable-model-invocation: true
---

# Update Changelog

Add one new section at the top of `CHANGELOG.md`. Invoke the skill with two parameters, previous version then current version:

`update-changelog $PREV_APP_VERSION $CURRENT_APP_VERSION`

Examples: `update-changelog v1.6.1 v1.7.0` or `update-changelog 1.6.1 1.7.0`.

If either version is missing, ask for both and stop. Do not invent a version and do not read `$PREV_APP_VERSION` from `CHANGELOG.md`.

Read [references/example.md](references/example.md) before writing and match its layout. The versions, issues, and notes in that file are mock data.

## Version

Accept `vMAJOR.MINOR.PATCH` or `MAJOR.MINOR.PATCH` for each argument. Reject anything else and ask again.

- `$PREV_APP_VERSION` is the first argument, with a leading `v` (`1.6.1` becomes `v1.6.1`).
- `$CURRENT_APP_VERSION` is the second argument, with a leading `v` (`1.7.0` becomes `v1.7.0`).
- If the top section is already `$CURRENT_APP_VERSION`, replace that section instead of inserting a duplicate.

`YYYY-MM-DD` is today's date. Do not leave a placeholder date.

## Section

You must add an additional section on top of the CHANGELOG.md. Place it immediately under the `# Changelog` title, above every older release. Leave one blank line after the title and one blank line after the new section.

This section must contain a header with this format: **[v$CURRENT_APP_VERSION](https://github.com/grafviktor/goto/compare/v$PREV_APP_VERSION...v$CURRENT_APP_VERSION) - YYYY-MM-DD**. Substitute the real version numbers and today's date. Keep the bold markers, the compare URL, and the spacing exactly as written.

Then you must add a list of tasks which was implemented since previous release, in the following format:

* Concise, clear, and formal description of the change [[details](https://github.com/grafviktor/goto/issues/$ISSUE_NUMBER)].
* Concise, clear, and formal description of the bugfix [[details](https://github.com/grafviktor/goto/issues/$ISSUE_NUMBER)].

One bullet per issue. Changes first, then bugfixes. Replace only the description wording with the real sentence. The description of a change or a bugfix must be a single sentence, at most 120 characters. Do not add a second sentence. Count only that sentence: do not count the leading `* ` or the details link. Keep `[[details](https://github.com/grafviktor/goto/issues/$ISSUE_NUMBER)].` exact, including the trailing period. A task with no issue number has no details link (`* Add rpm package support.`).

After the task list, add subsections only when they have bullets, in this order: `_Notes_`, then `_Breaking changes_`. Follow [references/example.md](references/example.md). Do not append either subsection after older releases. Omit a subsection when it has no bullets.

## What to include

Take every change for `$CURRENT_APP_VERSION` from this URL:

https://github.com/grafviktor/goto/issues?q=milestone%3AvX.Y.Z

`X.Y.Z` is the major, minor, and patch of `$CURRENT_APP_VERSION` (`v1.7.0` uses `milestone:v1.7.0`). One issue on that page is one changelog bullet. Include open and closed issues. Ignore pull requests.

```bash
gh issue list --repo grafviktor/goto --milestone "vX.Y.Z" --state all --limit 200 --json number,title,body,labels
```

When `gh` is unavailable:

```bash
curl -fsSL --get "https://api.github.com/search/issues" \
  --data-urlencode "q=repo:grafviktor/goto milestone:vX.Y.Z is:issue" \
  --data-urlencode "per_page=100"
```

Read each issue before writing. Also read the diff for that issue when the title is not enough to describe user-visible behavior, a break, or an edge case.

Write from the user's point of view, in the voice of the existing entries in `CHANGELOG.md`. State what changed in a single sentence. A bugfix starts with `Fix`. Do not copy issue titles, ticket prefixes, or words such as "in progress", "refactor", or "make linter happy". Skip an issue whose only changes are internal (tests, lint, formatting) and that has no user-visible effect.

A bugfix is an issue labeled bug, or a title prefixed with `BUG-` or `FIX-`, or a change that corrects broken behavior. Everything else in the milestone is a change.

## Notes

Put every note in the `_Notes_` subsection, after the task list and before `_Breaking changes_` when that subsection exists. A note is an edge case, a default, or other behavior the user must know that is not a breaking change.

Leave one blank line between the last task bullet and `_Notes_`, and one blank line between that heading and the first note. Keep the heading italic and unbolded. One bullet per note. Do not add a details link. Name the impact, what the user must do, and what still works. Omit the subsection when there are none.

## Breaking changes

Put every real breaking change in the `_Breaking changes_` subsection, after `_Notes_` when that subsection exists, and otherwise after the task list, before the blank line that closes the release. Follow the `_Breaking changes_` subsection in [references/example.md](references/example.md).

Leave one blank line before `_Breaking changes_` and one blank line between that heading and the first breaking-change bullet. Keep the heading italic and unbolded. One bullet per breaking change. Do not add a details link. Name the impact, what the user must do, and what still works. Omit the subsection when there are none.

Edit only `CHANGELOG.md`. Never commit.
