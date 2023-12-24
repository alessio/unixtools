#!/bin/sh

set -eo pipefail

if [ "$EUID" -ne 0 ]; then
  hdiutil attach -readonly -mountpoint $HOME/Library/Unixtools/ unixtools.dmg
  exit 0
fi

hdiutil attach -readonly -mountpoint /Library/Unixtools/ unixtools.dmg
