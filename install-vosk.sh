#!/bin/sh
version=0.3.42
vosk_file="vosk-linux-$(uname -m)-$version"
model=vosk-model-small-en-us-0.15
model_name=small-en-us

ok=1
! command -v unzip > /dev/null && echo 'you need unzip' && unset ok
! command -v wget > /dev/null && echo 'you need wget' && unset ok
[ "$ok" ] || exit

# not necessary but lets you run ./numen in this directory
ln -sf "/usr/local/share/vosk-models/$model_name" model

rm -rf tmp && mkdir -p tmp && cd tmp || exit

wget --no-verbose --show-progress "https://github.com/alphacep/vosk-api/releases/download/v${version}/$vosk_file.zip" || exit
unzip -q "$vosk_file" || exit
mkdir -p /lib && cp "$vosk_file/libvosk.so" /lib || exit
mkdir -p /usr/include && cp "$vosk_file/vosk_api.h" /usr/include || exit

wget --no-verbose --show-progress "https://alphacephei.com/kaldi/models/$model.zip" || exit
unzip -q "$model.zip" || exit
mkdir -p /usr/local/share/vosk-models || exit
rm -rf "/usr/local/share/vosk-models/$model_name" || exit
mv "$model" "/usr/local/share/vosk-models/$model_name" || exit

echo 'Installed successfully.'
