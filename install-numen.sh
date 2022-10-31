#!/bin/sh
# ./install-numen.sh [DESTDIR] [BINDIR]
#
# If the environment variable PACKAGING is set, compiling speech.go is left
# to you to do beforehand. It is meant for build systems packaging numen.
# Example: PACKAGING=true ./install-numen.sh "$DESTDIR" /usr/bin

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

	# not necessary but lets you run ./numen in this directory
	ln -sf /usr/share/vosk-models/small-en-us model
fi

# Install executables for internal use
libexec="$1/usr/libexec/numen"
rm -rf "$libexec" && mkdir -p "$libexec" || exit
cp awk instructor numen record scribe speech "$libexec" || exit
cp -r handlers "$libexec" || exit

# Install commands
bin="$1/${2:-/usr/local/bin}"
mkdir -p "$bin" || exit
cp wrapper "$bin/numen" || exit
# Install commands used in the default phrases
cp displaying "$bin" || exit

# Install the default phrases
mkdir -p /etc/numen && rm -rf /etc/numen/phrases && cp -r phrases /etc/numen || exit
# Install the manpage
scdoc < doc/numen.1.scd > /usr/share/man/man1/numen.1 || exit

echo 'Installed successfully.'
