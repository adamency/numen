#!/bin/sh
# A workaround to stop mawk buffering stdin.
if [ "$(command -v awk)" = /bin/mawk ]; then
	awk -Winteractive "$@"
else
	awk "$@"
fi
