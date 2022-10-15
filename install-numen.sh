#!/bin/sh

! command -v go > /dev/null && echo 'you need go (sometimes packaged as golang)' && exit 1

# Install the numen command
rm -rf /usr/libexec/numen && mkdir -p /usr/libexec/numen || exit
go build speech.go || exit
for f in *; do
	if [ -f "$f" ] && [ -x "$f" ]; then
		cp "$f" "/usr/libexec/numen/$f" || exit
	fi
done
cp -r handlers /usr/libexec/numen/handlers || exit
cat > /usr/local/bin/numen << 'EOF'
#!/bin/sh
cd /usr/libexec/numen
NUMEN_MODEL="${NUMEN_MODEL:-/usr/local/share/vosk-models/small-en-us}" exec ./numen "$@"
EOF
chmod +x /usr/local/bin/numen || exit

# Install the default phrases
mkdir -p /etc/numen && rm -rf /etc/numen/phrases && cp -r phrases /etc/numen || exit
# Install the displaying command used in the default phrases
cp displaying /usr/local/bin || exit
# Install the manpage if the scdoc command is installed
command -v scdoc > /dev/null && scdoc < numen.1.scd > /usr/share/man/man1/numen.1

echo 'Installed successfully.'
