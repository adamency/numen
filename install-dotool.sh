#!/bin/sh
version=1.1
distfile="https://git.sr.ht/~geb/dotool/archive/$version.tar.gz"
checksum=484bd2fac9f73c61b36b83086ce51fc69c60a99aa149e6e96fe413fa27747281

ok=1
! command -v go > /dev/null && echo 'you need go (sometimes packaged as golang)' && unset ok
! command -v tar > /dev/null && echo 'you need tar' && unset ok
! command -v wget > /dev/null && echo 'you need wget' && unset ok
[ "$ok" ] || exit

rm -rf tmp && mkdir -p tmp && cd tmp || exit

wget --no-verbose --show-progress -O dotool.tar.gz "$distfile" || exit
if [ "$(sha256sum dotool.tar.gz | cut -d' ' -f1)" != "$checksum" ]; then
	echo 'dotool.tar.gz did not match checksum'
	exit 1
fi

tar xf dotool.tar.gz || exit
cd "dotool-$version" || exit
DOTOOL_VERSION="$version" ./install.sh || exit

echo 'Installed successfully.'
