Set Shell bash
Type "gg -f temp -s mock_ssh_config"
Enter
Sleep 100ms

# Create a new host
Type "n"
Sleep 100ms
Type "network host"
Down
Ctrl+u
Type "127.0.0.1"
Down
Type "network host description"
Down
Type "yaml_config"
Ctrl+S
Sleep 100ms

# Copy host: network host (1)
Type "c"
Sleep 100ms

# Copy host: network host (2)
Type "c"
Sleep 100ms

# Change group to "yaml_config"
Type "z"
Sleep 100ms
Down
Enter

# Change group to "ssh_config"
Type "z"
Sleep 100ms
Down
Down
Enter

# Change group to "yaml_config"
Type "z"
Sleep 100ms
Down
Enter

# Select host: network host (2)
Down
Down

Type "q"