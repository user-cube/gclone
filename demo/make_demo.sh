#!/bin/bash
# Run demo_script and record
asciinema rec gclone.cast -c "./demo_script.sh" --overwrite

# Convert to GIF
asciicast2gif gclone.cast gclone.gif

# Optimize GIF
gifsicle -O3 --colors 256 gclone.gif -o gclone-demo.gif

echo "✅ Done! Your demo GIF is ready as gclone-demo.gif"