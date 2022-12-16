#!/bin/sh
version=1.0
distfile="https://git.sr.ht/~geb/dotool/archive/$version.tar.gz"
checksum=b73097f0c7be22e318e8ee446aed8291693a7198d335a82ca624a5887fe8d16d

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
./install.sh || exit

echo 'Installed successfully.'
