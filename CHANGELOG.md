# Changelog 

**[v1.4.0](https://github.com/grafviktor/goto/compare/v1.2.0...v1.3.0) - 2025-06-12**

* Able to display hosts  `~/.ssh/config` file [[details](https://github.com/grafviktor/goto/issues/45)].
* Various minor improvements and bugfixes.

**[v1.3.0](https://github.com/grafviktor/goto/compare/v1.2.0...v1.3.0) - 2025-02-25**

* Can now organize hosts into groups using `Group` input field [[details](https://github.com/grafviktor/goto/issues/66)].
* Resolve various issues on Windows platform; remove polling cycle which was used to track terminal size [[details](https://github.com/grafviktor/goto/issues/78)].
* Fix filtering mode issue when duplicate host appeared in the list [[details](https://github.com/grafviktor/goto/issues/84)].
* Various minor improvements and bugfixes.

**[v1.2.0](https://github.com/grafviktor/goto/compare/v1.1.0...v1.2.0) - 2024-09-30**

* Description field is now collapsible, press `v` to toggle the layout [[details](https://github.com/grafviktor/goto/issues/61)].
* You can invoke `ssh-copy-id` command for selected host by pressing `t` [[details](https://github.com/grafviktor/goto/issues/47)] _In test mode for Windows OS_.
* This release contains a fix for complex ssh commands containing quotes [[details](https://github.com/grafviktor/goto/issues/75)].
* Various minor improvements and bugfixes, including host focusing issues [[details](https://github.com/grafviktor/goto/issues/70)].

**[v1.1.0](https://github.com/grafviktor/goto/compare/v1.0.0...v1.1.0) - 2024-05-01**

* Support custom SSH parameters [[details](https://github.com/grafviktor/goto/issues/39)].
* Preserve host order when filter is enabled [[details](https://github.com/grafviktor/goto/issues/58)].
* Take default SSH parameters from system SSH configuration [[details](https://github.com/grafviktor/goto/issues/60)].
* Various minor improvements.

**[v1.0.0](https://github.com/grafviktor/goto/compare/v0.4.1...v1.0.0) - 2024-02-08**

* Adjust help menu, disable certain shortcuts when the operation is not available in current context [[details](https://github.com/grafviktor/goto/issues/43)].
* Disable automatic value copying from 'Title' to 'Host' input when the host is not new [[details](https://github.com/grafviktor/goto/issues/49)].
* Automate deb package building [[details](https://github.com/grafviktor/goto/issues/44)].
* Improve logging [[details](https://github.com/grafviktor/goto/issues/35)].

**[v0.4.1](https://github.com/grafviktor/goto/compare/v0.4.0...v0.4.1) - 2024-01-16**

* Fix an application title issue when filtering [[details](https://github.com/grafviktor/goto/issues/37)].
* Add rpm package support.

**[v0.4.0](https://github.com/grafviktor/goto/compare/v0.3.0...v0.4.0) - 2024-01-02**

* Display a confirmation dialog when delete a host from the database [[details](https://github.com/grafviktor/goto/pull/31)].
* Add validation rules for user input [[details](https://github.com/grafviktor/goto/pull/34)].
* When ssh throws an error user can see the full error message [[details](https://github.com/grafviktor/goto/pull/30)].
* Application title displays the ssh command which is going to be executed [[details](https://github.com/grafviktor/goto/pull/27)].

**[v0.3.0](https://github.com/grafviktor/goto/compare/v0.2.0...v0.3.0) - 2023-11-26**

* Fix a problem which led to broken cmd.exe UI on Windows platform [[details](https://github.com/grafviktor/goto/pull/14)].
* Introduce linter rules and add unit test coverage report.

**[v0.2.0](https://github.com/grafviktor/goto/compare/v0.1.2...v0.2.0) - 2023-11-11**

* The Application supports environment and command line parameters [[details](https://github.com/grafviktor/goto/issues/8)].
* Fix terminal resizing problem on Windows platform [[details](https://github.com/grafviktor/goto/issues/5)].

**[v0.1.2](https://github.com/grafviktor/goto/compare/v0.1.1...v0.1.2) - 2023-11-01**

* Resolve a problem with dissapearing host list when filter is enabled and a user is modifying the collection [[details](https://github.com/grafviktor/goto/issues/3)].

**v0.1.1 - 2023-10-30**

* Fix a focusing issue when saving an existing item using a different title [[details](https://github.com/grafviktor/goto/issues/1)].