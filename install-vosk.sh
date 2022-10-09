#!/bin/sh
version=0.3.42
vosk_file="vosk-linux-$(uname -m)-${version}"
model=vosk-model-small-en-us-0.15

! command -v unzip > /dev/null && echo 'you need unzip' && exit 1
! command -v wget > /dev/null && echo 'you need wget' && exit 1

wget "https://github.com/alphacep/vosk-api/releases/download/v${version}/${vosk_file}.zip" || exit
unzip "$vosk_file" || exit
cp "${vosk_file}/libvosk.so" /lib || exit
cp "${vosk_file}/vosk_api.h" /usr/include || exit

mkdir -p "/usr/local/share/vosk-models/$model" || exit
wget "https://alphacephei.com/kaldi/models/${model}.zip" || exit
unzip "${model}.zip" || exit
mv "$model" "/usr/local/share/vosk-models/small-en-us" || exit

echo 'Installed successfully.'
