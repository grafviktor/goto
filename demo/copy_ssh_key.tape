Set PlaybackSpeed 1
Output copy_ssh_key.gif
Require echo
Set Shell "bash"
Set FontSize 13
Set Width 600
Set Height 400
Set Padding 20

Show
# Run gg
Sleep 1s
Type "gg"
Sleep 1s
Enter
Sleep 1s

# Create a new host
Type "n"
Sleep 1s
# Set Title
Type "Remote machine"
Sleep 1s
Down
Sleep 1s

# Put a hostname
Ctrl+U
Sleep 1s
Type 'localhost'
Sleep 2s
Down

# Set Description field
Type 'ssh server'
Sleep 2s

# Save the host and exit
Ctrl+S
Sleep 3s

# Type 't' to copy ssh key and confirm the action by pressing 'y'
Type 't'
Sleep 2s
Type 'y'
# Now type your password and wait while the key is being installed
Sleep 1s
Type 'my-password'
Sleep 1s
Enter
Sleep 3s

# Key should be install now, let's exit to host list
Space
Sleep 2s

# Now let's try to connect to this host without password
Enter
Sleep 3s