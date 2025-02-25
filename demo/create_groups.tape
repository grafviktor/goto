Set PlaybackSpeed 1
Output create_groups.gif
Require echo
Set Shell "bash"
Set FontSize 13
Set Width 600
Set Height 400
Set Padding 20


# Run gg
Type "gg"
Sleep 300ms
Enter

# Modify host
Sleep 1s
Type "e"
Sleep 1s
# Set Group
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Type "Prod"
Sleep 1s
Ctrl+S
Sleep 1
Down

# Modify host
Sleep 1s
Type "e"
Sleep 1s
# Set Group
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Type "QA"
Sleep 1s
Ctrl+S
Sleep 1
Down

# Modify host
Sleep 1s
Type "e"
Sleep 1s
# Set Group
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Type "Prod"
Sleep 1s
Ctrl+S
Sleep 1
Down

# Modify host
Sleep 1s
Type "e"
Sleep 1s
# Set Group
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Type "QA"
Sleep 1s
Ctrl+S
Sleep 1

# Select group Prod
Type z
Sleep 1
Down
Sleep 1
Enter
Sleep 2

# Select group QA
Type z
Sleep 1
Down
Sleep 100ms
Down
Sleep 1
Enter
Sleep 2

# Select group All
Type z
Sleep 1
Up
Sleep 100ms
Up
Sleep 100ms
Enter
Sleep 1

# Show groups
Type v
Type v
Type v
Sleep 2
Escape
Sleep 1
Type y
Sleep 1