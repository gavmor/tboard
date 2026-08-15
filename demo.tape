# VHS tape file for generating tboard TUI demo GIF & screenshots
# Install VHS: brew install vhs (or go install github.com/charmbracelet/vhs@latest)

Output docs/images/tboard_demo.gif
Output docs/images/tboard_3d_waterfall.png

Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "GruvboxDark"
Set CornerRadius 10

# Launch tboard binary
Type "./bin/tboard 1.1.1.1 8.8.8.8 google.com github.com"
Enter
Sleep 3s

# Take screenshot of 3D Ridgeline Stack
Screenshot docs/images/tboard_3d_waterfall.png

# Cycle view to Split Table View
Type "v"
Sleep 3s

# Toggle theme to Gruvbox Light
Type "t"
Sleep 2s

# Quit
Type "q"
