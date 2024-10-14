# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Require:
#   Require <string>                Ensure a program is on the $PATH to proceed
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set LoopOffset <float>%         Set the starting frame offset for the GIF loop
#   Set Theme <json|string>         Set the theme of the terminal
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording
#   Set MarginFill <file|#000000>   Set the file or color the margin will be filled with.
#   Set Margin <number>             Set the size of the margin. Has no effect if MarginFill isn't set.
#   Set BorderRadius <number>       Set terminal border radius, in pixels.
#   Set WindowBar <string>          Set window bar type. (one of: Rings, RingsRight, Colorful, ColorfulRight)
#   Set WindowBarSize <number>      Set window bar size, in pixels. Default is 40.
#   Set TypingSpeed <time>          Set the typing speed of the terminal. Default is 50ms.
#
# Sleep:
#   Sleep <time>                    Sleep for a set amount of <time> in seconds
#
# Type:
#   Type[@<time>] "<characters>"    Type <characters> into the terminal with a
#                                   <time> delay between each character
#
# Keys:
#   Backspace[@<time>] [number]     Press the Backspace key
#   Down[@<time>] [number]          Press the Down key
#   Enter[@<time>] [number]         Press the Enter key
#   Space[@<time>] [number]         Press the Space key
#   Tab[@<time>] [number]           Press the Tab key
#   Left[@<time>] [number]          Press the Left Arrow key
#   Right[@<time>] [number]         Press the Right Arrow key
#   Up[@<time>] [number]            Press the Up Arrow key
#   Down[@<time>] [number]          Press the Down Arrow key
#   PageUp[@<time>] [number]        Press the Page Up key
#   PageDown[@<time>] [number]      Press the Page Down key
#   Ctrl+<key>                      Press the Control key + <key> (e.g. Ctrl+C)
#
# Display:
#   Hide                            Hide the subsequent commands from the output
#   Show                            Show the subsequent commands in the output

Set PlaybackSpeed 1
Output goto_and_tmux.gif
Require echo
Require tmux
Set Padding 5
Set FontSize 13
Set Width 600
Set Height 400
Set Padding 20

Show
Sleep 1

Hide
  Type "tmux -f /dev/null -L test new-session -- bash" Enter
  #Type "tmux split-window -d -h -- bash && \" Enter
  Type "tmux set status && \" Enter
  Type 'tmux setw pane-border-style "fg=0" && \' Enter
  Type 'tmux setw pane-active-border-style "fg=0"' Enter
  Type 'tmux bind x split-window -h -c "#{pane_current_path}"' Enter
  Sleep 0.5
  Type "export GG_HOME=. GG_LOG_LEVEL=debug" Enter
  Ctrl+L
  Sleep 1
Show

Type "gg"
Sleep 1s
Enter
Sleep 2s

# Create a new host
Type "n"
Sleep 1s
# Set Title
Type "localhost"
Sleep 1s
Down
Sleep 1s

# Replace localhost with 127.0.0.1 in Host input and add additional parameters
Ctrl+U
Sleep 1s
Type '127.0.0.1 '
Sleep 0.3s
Type '-t sh -c "tail -n 5 -f .colima/daemon.log"'
Sleep 2s
Down

# Set Description field
Type 'Tail kuber cluster logs on DEV box'
Sleep 2s

# Save the first host and exit
Ctrl+S
Sleep 3s

# Clone first host
Type c
Sleep 1s

# Focus on the copied host
Down
Sleep 1s

# Edit the copied host
Type e
Sleep 1s

# Delete " (1)" from the Title field
Backspace 4
Type " 2"
Sleep 1s

# Focus Host field and type new command
Down
Sleep 1s
Ctrl+U
Sleep 1s
Type "127.0.0.1 -t sh -c 'iostat -c5; sh read -p Done...'"
Sleep 2s

# Update description field
Down
Sleep 1s
Ctrl+U
Type "Display IO load"
Sleep 1s

# Save the second host and exit
Ctrl+S
Sleep 3s

Up

# Connect to the first host
Enter
Sleep 3s

# tmux bind to switch to next pane
Ctrl+B
Type x
# Sleep 0.5
# Ctrl+B
# Type o
Hide
  Type "export GG_HOME=. GG_LOG_LEVEL=debug" Enter
  Sleep 0.5
  Ctrl+L
  Sleep 1
Show

Sleep 1
Type "gg"
Sleep 1
Enter
Sleep 1
Down
Sleep 1
Enter
Sleep 5s
