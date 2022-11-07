#!/bin/sh
# A workaround to stop mawk buffering stdin.
case "$(realpath "$(command -v awk)")" in
	*/mawk) exec awk -Winteractive "$@";;
	*) exec awk "$@";;
esac
