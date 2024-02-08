Output duplicate_existing_record.gif
Require echo

Set Shell "bash"
Set FontSize 13
Set Width 600
Set Height 400
Show

Sleep 1s
Type "gg"
Sleep 1s
Enter
Sleep 2s

Type "c"
Sleep 1s
Type "c"
Sleep 1s

Down
Sleep 1s
Down
Sleep 1s

Up
Sleep 1s

# Edit host
Type "e"
Sleep 1s
# Append " which was cloned" to Server
Type " which was cloned"
Sleep 1s

# Go to description
Down
Sleep 1s
Backspace 11
Type remotedomain
Sleep 2s
Down
Sleep 1s
# Remove everything
Ctrl+U
Type "darkstar"
Sleep 2s
Down
Sleep 1s
Type "d.vader"
Sleep 2s

# Save
Ctrl+S

# Delete what was copied
Sleep 3s
Type d
Sleep 2s
Type y

Sleep 1s
Type d
Sleep 2s
Type y

Sleep 3s
Type "q"

Sleep 2s