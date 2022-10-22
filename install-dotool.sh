#!/bin/sh
version=1.0
distfile="https://git.sr.ht/~geb/dotool/archive/${version}.tar.gz"
checksum=22008f4ca55b50ce1f63020a52eca90a93a522e0a7ae2959cb185f44c3d6cd3c

ok=1
! command -v go > /dev/null && echo 'you need go (sometimes packaged as golang)' && unset ok
! command -v tar > /dev/null && echo 'you need tar' && unset ok
! command -v wget > /dev/null && echo 'you need wget' && unset ok
[ "$ok" ] || exit

mkdir -p tmp && cd tmp || exit

wget "$distfile" -O dotool.tar.gz || exit
if [ "$(sha256sum dotool.tar.gz | cut -d' ' -f1)" != "$checksum" ]; then
	echo 'dotool.tar.gz did not match checksum'
	exit 1
fi

tar xf dotool.tar.gz || exit
cd "dotool-${version}" || exit
./install.sh || exit

# Allow your user to run dotool without root permissions
echo KERNEL==\"uinput\", GROUP=\"$(logname)\", MODE:=\"0660\" > /etc/udev/rules.d/80-dotool-$(logname).rules
udevadm trigger

echo 'Installed successfully.'
