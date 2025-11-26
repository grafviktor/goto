# Test that we can set ssh_config path and it will survive
# application run in normal mode and overriding state file.

Set Shell bash

# Create ssh_config
Type "touch /tmp/delete_me_goto_ssh_config_test_file"
Enter
Sleep 100ms
Type "gg -f temp --set-ssh-config-path /tmp/delete_me_goto_ssh_config_test_file"
Enter
Sleep 100ms

# Run the app
Type "gg -f temp -l debug"
Enter
Sleep 100ms

# Create a random host
Type "n"
Sleep 100ms
Type "network host"
Down
Ctrl+u
Type "127.0.0.1"
Down
Type "network host description"
# Place the host to "manual hosts" group
Down
Type "manual hosts"
Ctrl+S
# Should always give some time when save a host
Sleep 100ms

Type "q"
Sleep 100ms
