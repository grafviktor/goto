# Display host list if ssh_config enabled, but there are not hosts in this file
Set Shell bash

# Open the app
Type "gg -f temp -s temp/ssh_config"
Enter
Sleep 100ms

# By default host list will be shown, so we can start adding
# a new host, without selecting the group first. If group list
# would be shown, then 'n' (new host) shortcut would be ignored.
Type "n"
Sleep 100ms

Type "network host 2"
Down
Ctrl+u
Type "127.0.0.2"
Down
Type "network host description 2"
Ctrl+S
# Should always give some time when save a host
Sleep 100ms
Type "q"