# GOTO - A simple SSH manager #

[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/grafviktor/goto/master/LICENSE)
[![Codecov](https://codecov.io/gh/grafviktor/goto/branch/develop/graph/badge.svg?token=tTyTsuCvNb)](https://codecov.io/gh/grafviktor/goto)

This tiny app helps to maintain a list of ssh servers. Unlike PuTTY it doesn't incorporate any connection logic, but relying on `ssh` utility which should be installed on your system.

Supported platforms: macOS, Linux, Windows.

## 1. Installation ##

### 1.1 Manual ###

* Download the latest version from the [Releases](https://github.com/grafviktor/goto/releases) section;
* Choose a binary file which matches your platform;
* Place the binary into your user's binary path;
* Optionally: rename `gg-${YOUR_PLATFORM_TYPE}` to `gg`.
* If you're on Linux or macOS, ensure that the binary has execution permissions:
  ```bash
  chmod +x gg
  ```

### 1.2 Using package manager ###

#### Debian or RedHat ####

RPM and DEB packages are available in the [releases](https://github.com/grafviktor/goto/releases/latest) section (these packages are not provided for pre-release builds).

#### Arch Linux (AUR) ####

_Maintained externally by the open-source community._

Install [goto-ssh-bin](https://aur.archlinux.org/packages/goto-ssh-bin) package. Also see the [build](https://aur.archlinux.org/cgit/aur.git/tree/PKGBUILD?h=goto-ssh-bin) file for additional details.

```bash
# Install goto
yay -S goto-ssh-bin
```

#### macOS (Homebrew) ####

_Maintained externally by the open-source community._

You can install `goto` via Homebrew using a community tap:

```bash
brew tap avasilic/goto
brew install goto-ssh
```

This installs the gg binary (renamed automatically). Run it with:
```bash
gg
```

## 2. Functional demo ##

### 2.1. Edit your database and connect to remote machines ###

![Shows how to open ssh session using goto](demo/edit_and_connect.gif)

### 2.2. Organize your hostnames into logical groups ###

![Shows how to switch between hosts groups](demo/switch_between_groups.gif)

### 2.3. Search efficiently across all your records ###

![Depicts how to search hosts through the database](demo/search_through_database.gif)

Find more demos and uses cases [here](demo/README.md).

## 3. Configuration ##

Please also refer [F.A.Q.](FAQ.md) page which provides additional configuration details and usage examples.

### 3.1. Command line options ###

* `-d` - disable feature, only supported value is ssh_config;
  ```bash
  gg -d "ssh_config" # since version 1.4.0
  ```
* `-e` - enable feature, only supported value is ssh_config;
  ```bash
  gg -e "ssh_config" # since version 1.4.0
  ```
* `-f` - specify the application home folder;
  ```bash
  gg -f /tmp/goto
  ```
* `-l` - log verbosity level. Only `info`(default) or `debug` values are currently supported;
  ```bash
  gg -l debug
  ```
* `-s` - define an alternative per-user SSH configuration file path;
  ```bash
  gg -s /mnt/nfs_share/ssh/config # since version 1.4.0
  ```
* `-h` - display help;
* `-v` - display version and configuration details.

### 3.2. Environment variables ###

* `GG_HOME` - specify the application home folder;
* `GG_LOG_LEVEL` - set log verbosity level. Only `info`(default) or `debug` values are currently supported.
* `GG_SSH_CONFIG_FILE_PATH` - define an alternative per-user SSH configuration file path.

## 4. File storage structure ##

2 file storages are supported:

* ssh_config - readonly storage type. Goto loads all hosts from your `~/.ssh/config` file. See `man ssh_config`, if you want to find out more about OpenSSH client configuration file.
* yaml file - writable storage type, but supports less options than ssh_config. Please section 4.1 if you want to find out more about yaml file structure and its location.

### 4.1 Yaml storage location and structure ###

You can only store your hosts in a yaml file, which is called `hosts.yaml`. The file is located in your user config folder which exact path depends on a running platform:

* on Linux, it's in `$XDG_CONFIG_HOME/goto` or `$HOME/.config/goto`;
* on Mac, it's in `$HOME/Library/Application Support/goto`;
* on Windows, it's in `%AppData%\goto`.

Usually you don't need to edit this file manually, but sometimes it's much more convenient to edit it with help of your favorite text editor, than using `goto` utility. The file structure is very simple and self-explanatory:

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

## 5. [F.A.Q.](FAQ.md) ##

## 6. [Contributing guidelines](CONTRIBUTING.md) ##

## 7. [Changelog](CHANGELOG.md) ##

## 8. [License](LICENSE) ##

## 9. Thanks ##

* To people who find time to contribute whether it is a bug report, a feature or a pull request.
* To [Charmbracelet project](https://charm.sh/) for the glamorous [Bubbletea](https://github.com/charmbracelet/bubbletea) library.
* To [JetBrains Team](https://www.jetbrains.com/) for their [support for Open-Source community](https://www.jetbrains.com/community/opensource/) and for the amazing products they make. That is a great boost indeed. I'm proudly placing their logo here as a humble "Thank You" gesture.

<div align="center">
  <a href="https://www.jetbrains.com/">
    <img
      height="40px"
      src="https://resources.jetbrains.com/storage/products/company/brand/logos/jetbrains.svg"
      alt="JetBrains logo."
    >
  </a>
</div>


