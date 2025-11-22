Set Shell bash
Type "gg -f temp -s mock_ssh_config -l debug"
Enter
Sleep 100ms
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