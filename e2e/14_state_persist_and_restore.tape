# In this test we run the app 2 times. Why do we need to run the app second time?
# Notice that that we didn't change any values during second start, we just ran the app and closed it at once.
# We needed it to read previously saved state. When we close the app, it calls state.Persist() method which overrides
# previously saved file. Therefore, if the app would fail to read some of the values during second run, the new state
# would be different.
#
# Overall, in this test we check persisting and reading of the following attributes:
# * group
# * screen_layout:
# * selected
# * enable_ssh_config
# * theme

Set Shell bash
Type "gg -f temp -d ssh_config"
Enter
Sleep 100ms

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
# Place the host to "manual hosts" group
Down
Type "manual hosts"
Ctrl+S
# Should always give some time when save a host
Sleep 100ms

# At this moment host with ID number 1 is selected (the host which we just created)
# Let's create another host
Type "n"
Sleep 100ms
Type "network host 2"
Down
Ctrl+u
Type "127.0.0.2"
Down
Type "network host description 2"
# Place the host to "manual hosts" group
Down
Type "manual hosts"
Ctrl+S
# Should always give some time when save a host
Sleep 100ms

# At this moment host with ID number 2 is selected
# Lets select "manual_hosts" group and switch to compact view
Type "z"
Down
Enter
Type "v"
Sleep 100ms
Type "v"
Sleep 100ms

# And now let's quit from the app
Type "q"

Type "gg -f temp --set-theme nord"
Enter

# Now run the app again
Sleep 100ms
Type "gg -f temp -s mock_ssh_config -l debug"
Enter
Sleep 100ms
# And quit from the app again
Type "q"
Sleep 100ms
