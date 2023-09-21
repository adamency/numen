#!/bin/sh
# ./get-dotool.sh ['install']
version=7688cc321c18420ac44abdb5d0acaba09af1675b
distfile="https://git.sr.ht/~geb/dotool/archive/$version.tar.gz"
checksum=6b331da5d52fc6f3c987fe7c85b6984c484eb43640eb940097143ff6523d11c9

if [ "$*" != '' ] && [ "$*" != install ]; then
	echo bad usage
	exit 1
fi

ok=1
! command -v go >/dev/null && echo 'you need go (aka golang)' && unset ok
! command -v tar >/dev/null && echo 'you need tar' && unset ok
! command -v scdoc >/dev/null && echo 'you need scdoc' && unset ok
! command -v wget >/dev/null && echo 'you need wget' && unset ok
! test -d /usr/include/xkbcommon && echo 'you need libxkbcommon-dev' && unset ok
[ "$ok" ] || exit

export DOTOOL_VERSION="$version"

mkdir -p tmp && cd tmp
if ! [ "$1" ]; then
	wget --no-verbose -O dotool.tar.gz "$distfile" || exit
	if [ "$(sha256sum dotool.tar.gz | cut -d' ' -f1)" != "$checksum" ]; then
		echo dotool.tar.gz did not match checksum
		exit 1
	fi
	tar xf dotool.tar.gz || exit
	cd "dotool-$version" || exit
	./build.sh || exit
else
	cd "dotool-$version" || exit
	./build.sh install || exit
fi
