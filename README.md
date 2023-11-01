# GOTO - Simple SSH manager #

This utility helps to maintain a list of ssh servers. Unlike PuTTY id doesn't incorporate any connection logic, but relying on `ssh` utility which should be installed on your system.

## Installation ##

* Download the zip file from the releases section;
* Choose a binary file which matches your platform;
* Place the binary into your user's binary path;
* Optionally: rename `gg-$platform` into `gg`.

## Functional preview ##

### Edit and connect to a remote box ###

![Small demo where we open ssh session using goto](demo/edit_and_connect.gif)

### Duplicate an existing record ###

![Small demo where duplicate an existing record in goto database](demo/duplicate_existing_record.gif)

### Find a requried host easily among all your records ###

![Small demo where we open ssh session using goto](demo/search_through_database.gif)

## File storage structure ##

Currently you can only store your hosts in a yaml file, which is called `hosts.yaml`. The file is located in your user config folder which exact path depends on a running platform:

* On Linux, it's in `$XDG_CONFIG_HOME/goto` or `$HOME/.config/goto`;
* On Mac, it's in `$HOME/Library/Application Support/goto`;
* On Windows, it's in `%AppData%\goto`.

Usually you don't need to edit this file manually, but sometimes it's much convenient to edit it into your favorite text editor, than using `goto` utility. The file structure is very simple and self-explanatory:

```yaml
- host:
    title: kernel.org
    description: Server 1
    address: 127.0.0.1
- host:
    title: microsoft.com
    description: Server 2
    address: 127.0.0.1
    network_port: 22
    username: satya
    identity_file_path: /home/user/.ssh/id_rsa_microsoft
```

## Changelog ##

**v0.1.1**

Fix a focusing issue when saving an existing item using a different title.

## Known bugs ##

* Terminal resizing in Windows OS is not yet supported, as `cmd.exe` does not fire window resize events;
* You cannot disable generating `debug.log` file;
* User input validators do not exist;
* There is no confirmation dialog when you delete an existing item from the database;
* Maybe some other things as the utility hasn't even reached a stable version.

## License ##

[MIT](LICENSE)