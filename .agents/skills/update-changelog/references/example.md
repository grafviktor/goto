# update-changelog example

Invocation:

```text
update-changelog v1.2.0 v1.3.0
```

Equivalent: `update-changelog 1.2.0 1.3.0`.

Result: this section is inserted at the top of `CHANGELOG.md`, directly under `# Changelog`, with one blank line before the next release.

```markdown
# Changelog

**[v1.3.0](https://github.com/grafviktor/goto/compare/v1.2.0...v1.3.0) - 2026-11-15**

* Add an export command for the host list [[details](https://github.com/grafviktor/goto/issues/10)].
* Ask for confirmation before discarding unsaved edits [[details](https://github.com/grafviktor/goto/issues/11)].
* Fix a crash when the host list is empty [[details](https://github.com/grafviktor/goto/issues/12)].

_Notes_

* Mouse support is disabled by default. Enable it with `goto -e mouse` or in the application settings.
* The embedded terminal is enabled by default, so SSH sessions now run inside the application. To restore the previous behavior, run `goto -d embedded_terminal`.

_Breaking changes_

* The default screen layout is now compact. Pass `--layout full` to restore the previous layout.
* The state file key `theme_name` was renamed to `theme`. The saved theme is not restored after the upgrade; other settings keep working.

**[v1.2.0](https://github.com/grafviktor/goto/compare/v1.1.0...v1.2.0) - 2026-06-08**

* Remember the last selected host between restarts [[details](https://github.com/grafviktor/goto/issues/9)].
```
