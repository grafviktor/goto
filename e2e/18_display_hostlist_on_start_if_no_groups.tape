# Display host list if there are no groups
Set Shell bash

# Disable ssh_config
Type "gg -f temp -d ssh_config"
Enter
Sleep 100ms

# Open the app
Type "gg -f temp"
Enter
Sleep 100ms

# By default host list will be shown, so we can start adding
# a new host, without selecting the group first. If group list
# would be shown, then 'n' (new host) shortcut would be ignored.
Type "n"
Sleep 100ms

Type "network host"
Down
Ctrl+u
Type "127.0.0.1"
Down
Type "network host description"
Ctrl+S
# Should always give some time when save a host
Sleep 100ms
Type "q"