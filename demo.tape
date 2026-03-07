Output demo.gif

Set Shell zsh
Set FontSize 14
Set Width 800
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"

# Set PATH silently
Hide
Type "export PATH=$HOME/go/bin:$PATH"
Enter
Sleep 0.1s
Type "alias frefresh='frefresh --demo'"
Enter
Sleep 0.1s
Type "clear"
Enter
Sleep 0.3s
Show

# Start frefresh in demo mode
Type "frefresh"
Sleep 0.5s
Enter
Sleep 2s

# Main menu: select "Refresh tables" (option 1)
Type "1"
Sleep 1.5s

# Select customer: Contoso (option 1)
Type "1"
Sleep 1.5s

# Select environment: DEV (option 1)
Type "1"
Sleep 1.5s

# Select semantic model: HR (option 1)
Type "1"
Sleep 1.5s

# Table checkbox: toggle "All" with space to select everything
Type " "
Sleep 0.5s

# Confirm with enter
Enter
Sleep 1.5s

# Confirm refresh: y
Type "y"
Enter
Sleep 6s

# Show the result
Sleep 3s

# Quit
Type "q"
Sleep 1s
