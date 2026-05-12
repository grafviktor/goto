Set Shell bash
Type "gg -f temp -s 11_ssh_config_weird_format.config"
Enter
Sleep 100ms

# Change group from "all" to "ssh_config"
Type z
Down
Enter

# Go to the very end of the list and exit the application.
# After that we check that all hosts were read from the config file
# by comparing selected host ID in state file and the quantity of hosts in ssh config.
Type G

Type "q"