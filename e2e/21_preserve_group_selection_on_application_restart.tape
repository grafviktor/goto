# Ensure that when we persist group name in state file application restart.
Set Shell bash

# Run the app
Type "gg -f temp"
Enter
Sleep 100ms

# There are 3 groups in the list, apart from "~ all ~". "group 3" is preselected in the initial state file. Switching from group "group 3" to group with name "group 1"
Up
Up
Enter
Sleep 100ms

# Now quit the application. In the expected state file selected group name should be "group 1".
Type "q"