# Ensure that when we open the app and there is more than one group, we must display group list.
Set Shell bash

# Disable ssh_config for simplicity - there are 3 additional
# hosts in mock_ssh_config and I do not want to calculate their indices
# for determining proper selected_host_index in expected state file.
Type "gg -f temp -d ssh_config"
Enter
Sleep 100ms

# Run the app
Type "gg -f temp"
Enter
Sleep 100ms

# Assuming that we see group list. Moving down from group "~ all ~" to group with name "production"
Down
Down
Down
Enter
Sleep 100ms

# Now let's create a new host. If host list would be shown on application start,
# we would not be able to select "production" group. Note, that we don't have
# to specify "production" group explicitly - It will be taken from current group.
Type "n"
Sleep 100ms
Type "production network host 3"
Down
Ctrl+u
Type "127.0.0.3"
Down
Type "production network host description 3"
Ctrl+S
# Should always give some time when save a host
Sleep 100ms
Type "q"