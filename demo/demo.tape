# genius — README demo tape (https://github.com/charmbracelet/vhs)
#
# Records a deterministic tour of the TUI against a synthetic, self-contained
# study workspace (demo/home) so no generation engine is ever invoked — every
# guide and Q&A opened here already exists on disk.
#
# Regenerate the GIF from the repo root:
#   vhs demo/demo.tape
#
# Requires: vhs, ttyd, ffmpeg on PATH, and a built ./genius binary
#   (go build -o genius .).

Output demo/genius.gif

Set Shell bash
Set FontSize 18
Set Width 1280
Set Height 720
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 55ms

# --- setup: point genius at the demo workspace, use the freshly built binary ---
Hide
Type "export GENIUS_HOME=$PWD/demo/home PATH=$PWD:$PATH"
Enter
Type "clear"
Enter
Show

# --- home dashboard: wordmark banner + course list with g·N q·N chips ---
Type "genius"
Enter
Sleep 2s
Down@600ms 1
Sleep 900ms
Up@600ms 1
Sleep 700ms

# --- open the study guide (already on disk → Glamour-rendered, no engine) ---
Enter
Sleep 1.6s
Space
Sleep 1.3s
Escape
Sleep 700ms

# --- revise: r opens the Q&A picker (whole / chapter / merged), then the quiz ---
Type "r"
Sleep 1.2s
Enter
Sleep 1.1s

# type an answer, reveal, self-grade "knew it"
Type@70ms "an upper bound on cost growth"
Sleep 500ms
Enter
Sleep 1.9s
Type "y"
Sleep 1.3s
Escape
Sleep 700ms

# --- chapter hub: scope guides / Q&A to selected chapters (f) ---
Type "f"
Sleep 1.4s
Space
Sleep 500ms
Down@500ms 2
Space
Sleep 1.3s
Escape
Sleep 700ms

# --- keybinding help overlay (?) ---
Type "?"
Sleep 2s
Escape
Sleep 600ms

# --- quit ---
Ctrl+C
Sleep 700ms
