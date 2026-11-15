# generate-release-notes example

Invocation:

```text
generate-release-notes v1.3.0
```

Equivalent: `generate-release-notes 1.3.0`.

Result: `RELEASE_NOTES_v1.3.0.md` at the repository root. The versions, issues, and notes below are mock data.

```markdown
Hosts can now be exported, and unsaved edits ask for confirmation before they are discarded.

## What's changed ##

### Improvements ###

* Add an export command for the host list [[details](https://github.com/grafviktor/goto/issues/10)].
* Ask for confirmation before discarding unsaved edits [[details](https://github.com/grafviktor/goto/issues/11)].

### Fixes ###

* Fix a crash when the host list is empty [[details](https://github.com/grafviktor/goto/issues/12)].

### Technical ###

* Build release packages with GoReleaser [[details](https://github.com/grafviktor/goto/issues/13)].

## Notes ##

* Mouse support is disabled by default. Enable it with `goto -e mouse`. Keyboard navigation is unchanged.

## Breaking changes ##

The default screen layout is now compact. Pass `--layout full` to restore the previous layout. Saved hosts are unchanged.

```bash
goto --layout full
```

<hr>
```

Omit `### Improvements ###`, `### Fixes ###`, `### Technical ###`, `## Notes ##`, or `## Breaking changes ##` when that part has nothing to say.
