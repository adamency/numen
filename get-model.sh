#!/bin/sh
# ./get-model.sh ['install']
f=vosk-model-small-en-us-0.15
name=small-en-us
checksum=30f26242c4eb449f948e42cb302dd7a686cb29a3423a8367f99ff41780942498

if [ "$*" != '' ] && [ "$*" != install ]; then
	echo bad usage
	exit 1
fi

ok=1
! command -v unzip >/dev/null && echo 'you need unzip' && unset ok
! command -v wget >/dev/null && echo 'you need wget' && unset ok
[ "$ok" ] || exit

mkdir -p tmp && cd tmp
if ! [ "$1" ]; then
	wget --no-verbose -O "$f.zip" "https://alphacephei.com/kaldi/models/$f.zip" || exit
	if [ "$(sha256sum "$f.zip" | cut -d' ' -f1)" != "$checksum" ]; then
		printf %s\\n "$f.zip did not match the checksum"
		exit 1
	fi
	echo Downloaded successfully.
else
	unzip -qo "$f.zip" || exit
	mkdir -p /usr/share/vosk-models || exit
	mv -T "$f" "/usr/share/vosk-models/$name" || exit
	echo Installed successfully.
fi
