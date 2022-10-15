#!/bin/sh
version=1.0.1
distfile="https://github.com/ReimuNotMoe/ydotool/archive/v${version}.tar.gz"
checksum=02311cdc608f205711b06a95e5fd71093b2294f4920efc526f5e98a2ddab42b8

ok=1
! command -v cmake > /dev/null && echo 'you need cmake' && unset ok
! command -v g++ > /dev/null && echo 'you need g++' && unset ok
! command -v make > /dev/null && echo 'you need make' && unset ok
! command -v scdoc > /dev/null && echo 'you need scdoc' && unset ok
! command -v tar > /dev/null && echo 'you need tar' && unset ok
! command -v wget > /dev/null && echo 'you need wget' && unset ok
[ "$ok" ] || exit

# Directory for the socket. See ./ydotool-files/socket-group-permission.patch
mkdir -p /var/lib/ydotoold || exit
chown :input /var/lib/ydotoold && chmod 775 /var/lib/ydotoold || exit

cp ydotool-files/80-uinput.rules /usr/lib/udev/rules.d || exit
cp ydotool-files/50-ydotool.conf /usr/share/X11/xorg.conf.d || exit

mkdir -p tmp && cd tmp || exit

wget "$distfile" -O ydotool.tar.gz || exit
if [ "$(sha256sum ydotool.tar.gz | cut -d' ' -f1)" != "$checksum" ]; then
	echo 'ydotool.tar.gz did not match checksum'
	exit 1
fi

tar xf ydotool.tar.gz || exit
cd "ydotool-${version}" || exit
patch -Np1 < ../../ydotool-files/man-update-for-1.0.1.patch || exit
patch -Np1 < ../../ydotool-files/socket-group-permission.patch || exit
rm -rf build && mkdir build && cd build || exit
cmake .. && make -j "$(nproc)" || exit
cp ydotool ydotoold /usr/bin || exit

echo 'Installed successfully.'
echo 'You will need to add yourself to group input:

    $ sudo usermod -a -G input $USER

and then re-login to make it effective.'
