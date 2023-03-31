#!/bin/sh
version=90184107489abb7a440bf1f8df9b123acc8f9628
distfile="https://git.sr.ht/~geb/dotool/archive/$version.tar.gz"
checksum=9f5ff7513307f0829b75f43a3fd0f3929747cddcc83904ee066fc578b4b15492

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
