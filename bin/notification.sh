if [ "$(uname)" = "Darwin" ]; then
  afplay /System/Library/Sounds/Glass.aiff
else
  powershell -c "[System.Media.SystemSounds]::Hand.Play()"
fi