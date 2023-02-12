# Numen

Numen is voice control for desktop computing without a keyboard or mouse.
It works system-wide on your Linux machine.

A short demonstration can be found on:
[https://numenvoice.com](https://numenvoice.com)

## Install From Source

`go` (sometimes packaged as `golang`) is required.

The [speech recognition library](https://alphacephei.com/vosk) and an English
model (about 40MB) can be installed with:

    sudo ./install-vosk.sh && sudo ./install-model.sh

The [dotool](https://sr.ht/~geb/dotool) command which is used to simulate
input, can be installed with:

    sudo ./install-dotool.sh

Finally, `numen` itself can be installed with:

    sudo ./install-numen.sh

## Permission

`dotool` requires permission to `/dev/uinput` to create the virtual input
devices, and the udev rule grants this to users in group input.

You could check the output of `groups` or just try:

    echo type hello | dotool

If need be, you can run:

    sudo usermod -a -G input $USER

and re-login and trigger the udev rule, or just reboot.

## Getting Started

Once you've got a microphone, you can run it with: `numen`<br> You should
be able to type "hey" by saying "hoof eve yank", and transcribe a sentence
after saying "scribe".  You can terminate it by pressing Ctrl+c or saying
"troll cap".

If nothing happened, you might need to specify the right microphone with the
`--mic` option.  See `numen --list-mics` for what's available.

The default phrases are in the `/etc/numen/phrases/` directory and I'd
start by looking at `character.phrases` with the alphabet and symbols, and
`control.phrases` with the modifiers, backspace and friends.  Have a go in
your text editor.

## Going Further

Voice control makes an efficient keyboard but a wack mouse.  At first I
thought I'd need something like eye tracking, but now I just use keyboard
based programs, which are thankfully the most productive kind of program.
The main two are [Neovim](https://neovim.io), my text editor, and
[qutebrowser](https://qutebrowser.org), my web browser.

I've also made a desktop environment that works well with voice control,
called [tiles](https://git.sr.ht/~geb/tiles).

## Contact

You can ask for help or send patches by composing an email to
[~geb/public-inbox@lists.sr.ht](https://lists.sr.ht/~geb/public-inbox).
You're also welcome to join our Matrix chat at
[#numen:matrix.org](https://matrix.to/#/#numen:matrix.org).

## Support Me

[Thank you!](https://liberapay.com/geb)

## License

GPLv3 only, see [LICENSE](./LICENSE).

Copyright (c) 2022-2023 John Gebbie
