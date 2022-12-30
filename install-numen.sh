#!/bin/sh
# ./install-numen.sh [DESTDIR] [BINDIR]
#
# If the environment variable PACKAGING is set, compiling speech.go is left
# to you to do beforehand. It is meant for build systems packaging numen.
# Example: PACKAGING=true ./install-numen.sh "$DESTDIR" /usr/bin
version="$(git describe --long --abbrev=12 --tags --dirty 2>/dev/null || echo 0.5)"

if ! [ "$PACKAGING" ]; then
	ok=1
	! command -v arecord > /dev/null && echo 'you need the alsa-utils package' && unset ok
	! command -v dmenu > /dev/null && echo 'you need dmenu' && unset ok
	! command -v gcc > /dev/null && echo 'you need gcc' && unset ok
	! command -v go > /dev/null && echo 'you need go (sometimes packaged as golang)' && unset ok
	! command -v scdoc > /dev/null && echo 'you need scdoc' && unset ok
	! command -v xdotool > /dev/null && echo 'you need xdotool' && unset ok
	! command -v xset > /dev/null && echo 'you need xset' && unset ok
	[ "$ok" ] || exit

	go build speech.go || exit
fi

# Install executables for internal use
libexec="$1/usr/libexec/numen"
rm -rf "$libexec" && mkdir -p "$libexec" || exit
cp awk instructor numen record scribe speech "$libexec" || exit
cp -r handlers "$libexec" || exit

# Install commands
bin="$1/${2:-/usr/local/bin}"
mkdir -p "$bin" || exit
sed "1 a \
export NUMEN_VERSION=$version" wrapper > "$bin/numen" && chmod +x "$bin/numen" || exit
# Install commands used in the default phrases
cp displaying "$bin" || exit

# Install the default phrases
mkdir -p "$1/etc/numen" && rm -rf "$1/etc/numen/phrases" && cp -r phrases "$1/etc/numen" || exit
# Install the manpage
mkdir -p "$1/usr/share/man/man1" && scdoc < doc/numen.1.scd > "$1/usr/share/man/man1/numen.1" || exit

echo 'Installed successfully.'
