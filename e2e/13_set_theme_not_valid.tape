Set Shell bash
# there is no such theme, the next line should give an error and state file will not be created
Type "gg -f temp --set-theme no_such_theme"
Enter
# ensure that config does not contain invalid theme name
# we must run and exit the app to create state file, otherwise the test will fail
Sleep 1
Type "gg -f temp"
Enter
Sleep 100ms
Type "q"
Enter