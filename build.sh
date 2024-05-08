#!/bin/sh
# ./build.sh ['install']
: "${NUMEN_VERSION=$(git describe --long --abbrev=12 --tags --dirty 2>/dev/null || echo 0.7)}"
: "${NUMEN_DESTDIR=}"
: "${NUMEN_BINDIR=usr/local/bin}"
: "${NUMEN_DEFAULT_PHRASES_DIR=/usr/share/numen/phrases}"
: "${NUMEN_SCRIPTS_DIR=/usr/share/numen/scripts}"
: "${NUMEN_DEFAULT_MODEL_PACKAGE=vosk-model-small-en-us}"
: "${NUMEN_DEFAULT_MODEL=/usr/share/vosk-models/small-en-us}"

if [ "$*" != '' ] && [ "$*" != install ]; then
	echo bad usage
	exit 1
fi

if ! [ "$NUMEN_SKIP_CHECKS" ]; then
	ok=1
	! command -v arecord >/dev/null && echo 'you need alsa-utils' && unset ok
	! command -v dotool >/dev/null && echo 'you need dotool' && unset ok
	! command -v gcc >/dev/null && echo 'you need gcc' && unset ok
	! command -v go >/dev/null && echo 'you need go (aka golang)' && unset ok
	! command -v scdoc >/dev/null && echo 'you need scdoc' && unset ok
	[ "$ok" ] || exit
fi

if ! [ "$1" ]; then
	if [ "$NUMEN_DEFAULT_MODEL_PATHS" ]; then
		echo 'NOTE $NUMEN_DEFAULT_MODEL_PATHS is deprecated, favoring $NUMEN_DEFAULT_MODEL.'
	fi

	go build -ldflags "-X 'main.Version=$NUMEN_VERSION'
		-X 'main.DefaultModelPackage=$NUMEN_DEFAULT_MODEL_PACKAGE'
		-X 'main.DefaultModel=$NUMEN_DEFAULT_MODEL'
		-X 'main.DefaultPhrasesDir=$NUMEN_DEFAULT_PHRASES_DIR'" || exit
	echo Built successfully.
else
	install -Dm755 numen numenc -t "$NUMEN_DESTDIR/$NUMEN_BINDIR" || exit
	install -Dm755 scripts/* -t "$NUMEN_DESTDIR/$NUMEN_SCRIPTS_DIR" || exit
	install -Dm644 phrases/* -t "$NUMEN_DESTDIR/$NUMEN_DEFAULT_PHRASES_DIR" || exit
	sed -i "s:/usr/share/numen/scripts:$NUMEN_SCRIPTS_DIR:g" \
		"$NUMEN_DESTDIR/$NUMEN_SCRIPTS_DIR"/* \
		"$NUMEN_DESTDIR/$NUMEN_DEFAULT_PHRASES_DIR"/* || exit
	mkdir -p "$NUMEN_DESTDIR/usr/share/man/man1" || exit
	scdoc < doc/numen.1.scd > "$NUMEN_DESTDIR/usr/share/man/man1/numen.1" || exit
	echo Installed Successfully.

	if [ -e /usr/local/share/vosk-models/small-en-us ]; then
		echo '
NOTE /usr/local/share/vosk-models/small-en-us is deprecated.
To move it:
mv -T /usr/local/share/vosk-models/small-en-us /usr/share/vosk-models/small-en-us'
	fi
fi
