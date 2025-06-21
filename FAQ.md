# F.A.Q. #

## General questions ##

**Q: Where can I find the list of supported shortcuts?**

Press question mark (`?`) when you're in the application. You will see all supported key combinations.

**Q: Why when I filter my hosts, I see titles which do not match my filter criteria?**

This is because filter checks Title and Description fields of your hosts.

**Q: How can I use ssh jump hosts or any other sophisticated configuration?**

In general I would advise you to put this configuration into your ssh_config file (this file is usually located in `$HOME/.ssh/config`) and enable ssh_config support in the app. Another option is to insert your ssh command right into the Address field of a specific host, though it'll look a bit messy, it will work.

**Q: When I add a new host, the default identity file is set to id_rsa. However I want to use ED25519 key instead. How can I do that?**

The application fully relies on your current ssh settings. Please execute `ssh -G yourhostname` command and check the first identity file name. This is exactly what you get in GoTo.
```bash
ssh -G yourhostname | grep identityfile -m 1
# => identityfile ~/.ssh/id_rsa
```
To change that, update your `$HOME/.ssh/config` to something like below:
```
Host *
  IdentityFile ~/.ssh/id_ed25519
  IdentityFile ~/.ssh/id_rsa
```

**Q: I don't like typing password every time I connect to my ssh server. How can I save the passwords?**

You cannot. The author of this application is not a security expert and can't take the risk of storing someone's sensitive data. You can tweak the app and integrate `sshpass` if you wish. But I would recommend relying on ssh identity files instead.

**Q: Are there any viruses or trojans in the app?**

There are not. What the app does - it builds the list of hosts based on your ssh_config and hosts.yaml files. Essentially it's just a wrapper around `ssh` command. Once you hit enter on a certain host, the app transfers your session control to another process which is your ssh utility. However, GoTo does read ssh process output for logging. You can check what's exactly read by running the command in debug mode and checking the app.log file.

The app does not send or request anything from the network.

**Q: I do not want to use binary builds you provided. How can build the app myself?**

It's very easy. Make sure that `go` and `make` are installed in your system. Just clone the repo and and run `make build`. I also suggest you to read through the Makefile, it's well commented and self-explanatory.

## OpenSSH client configuration file support (only version 1.4.0 and above) ##

**Q: Why was ssh_config support added?**

GoTo keeps all hosts in a yaml file. The number of supported options in yaml is limited and it would be difficult to extend the application to support every option which is available in ssh_config. Since release 1.4.0 GoTo supports 2 storage types - yaml file and ssh_config (which is your `$HOME/.ssh/config`). When the application starts, it reads hosts from both files and displays these hosts in the UI.

**Q: I don't like the app reading my `~/.ssh/config`. How can I disable that?**

Just execute the command below to disable this feature. The next time you run the app, it will not read your ssh_config.
```bash
gg -d ssh_config
```
To re-enable this feature, execute:
```bash
gg -e ssh_config
```

**Q: How can I manually set path to my ssh_config file?**

You have 2 options:
1. Use `-s` key to define ssh_config file location. The downside of this approach is that you will need to use this key every time you run the app.
2. More convenient way is to define environment variable with name: `GG_SSH_CONFIG_FILE_PATH`. For instance:
    ```bash
    export GG_SSH_CONFIG_FILE_PATH="/mnt/c/Users/D.Vader/.ssh/config"
    ```

**Q: Can I assign a specific group to a host which is located in my ssh_config?**

Yes, you can! By default all hosts loaded from `$HOME/.ssh/config` will be put to a group with name `ssh_config`. However, you can assign them any group you want. Just add meta comments into your host's definition block. Currently only 2 meta comments are supported: `GG:GROUP` and `GG:DESCRIPTION`. Here is an example:

```
Host SOME_HOST_ALIAS
  # GG:GROUP Production
  # GG:DESCRIPTION Shadow Prod Machine
  User appuser
  HostName production2.intranet
  Port 2222
  IdentityFile /home/user/.ssh/id_rsa_prod
```

Once the application runs, you will find your host in group with name "Production" (press `z` to display the group list). The description will be "Shadow Prod machine".

**Q: Why can't I edit hosts, which were loaded from `~/.ssh/config`?**

AI says that there are over 100 supported configuration options in ssh_config. The UI which could support all these options would look rather complex. If you want to change hosts loaded from your ssh config, edit your ssh_config file directly using your favorite text editor and restart `goto`, to update the changes.
