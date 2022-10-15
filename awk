#!/bin/sh
# A workaround to stop mawk buffering stdin.
case "$(realpath "$(command -v awk)")" in
	*/mawk) awk -Winteractive "$@";;
	*) awk "$@";;
esac
