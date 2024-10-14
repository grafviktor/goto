Output duplicate_existing_record.gif
Require echo

Set Shell "bash"
Set FontSize 13
Set Width 600
Set Height 400
Set Padding 20

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
# Rename to "The Death Star"
Backspace 12
Type "The Death Star"
Sleep 1s

# Go to description
Down
Sleep 1s
Backspace 11
Type "d.vader"
Sleep 300ms
Type "@"
Sleep 300ms
Type "death-star.space"
Sleep 300ms
Type " -p 2222"
Sleep 3s
Down
Sleep 1s
# Remove everything
Ctrl+U
Type "turbolaser cannon secure shell"
Sleep 2s

# Save
Ctrl+S
Sleep 3s

# Delete other hosts
Up
Sleep 1s
Type d
Sleep 1s
Type y

Up
Sleep 1s
Type d
Sleep 1s
Type y

Sleep 3s
Type "q"

Sleep 2s