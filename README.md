# Numen

Numen is voice control for computing without a keyboard or mouse,
and works system-wide on your Linux machine.

There's a short demonstration on:
[https://numenvoice.org](https://numenvoice.org)

## Install From Source

`go` is required. (It's sometimes packaged as `golang`)

The [speech recognition library](https://alphacephei.com/vosk) and an English
model (about 40MB) can be installed with:

    sudo ./install-vosk.sh && sudo ./install-model.sh

The [dotool](https://sr.ht/~geb/dotool) command which simulates the input,
can be installed with:

    sudo ./install-dotool.sh

Finally, `numen` itself can be installed with:

    sudo ./install-numen.sh

## Permission

`dotool` requires permission to `/dev/uinput` to create the virtual input
devices, and a udev rule grants this to users in group input.

You could try:

    echo type hello | dotool

and if need be, you can run:

    sudo groupadd -f input
    sudo usermod -a -G input $USER

and re-login and trigger the udev rule or just reboot.

## Getting Started

Once you've got a microphone, you can run it with:

    numen

There shouldn't be any output but you should be able to type "hey" by saying
"hoof eve yank" and transcribe a sentence after saying "scribe".  You can
terminate it by pressing Ctrl+c which is "troll cap".

If nothing happened, check it's using the right audio device with:

    timeout 5 numen --verbose --audiolog=me.wav
    aplay me.wav

and specify a `--mic` from `--list-mics` if not.

Now you can have a go in your text editor, the default phrases are in the
`/etc/numen/phrases` directory.

## Going Further

I use numen for all my computing and stick to keyboard-based programs like
[Neovim](https://neovim.io) and [qutebrowser](https://qutebrowser.org), my text
editor and browser.  I also use a minimal desktop environment I wrote called
[Tiles](https://git.sr.ht/~geb/tiles) that doesn't require a pointer device.

## Contact and Matrix Chat

You can send questions, thoughts or patches by composing an email to
[~geb/public-inbox@lists.sr.ht](https://lists.sr.ht/~geb/public-inbox).

You're also welcome to join our Matrix chat at
[#numen:matrix.org](https://matrix.to/#/#numen:matrix.org).

## See Also

* [Tiles](https://git.sr.ht/~geb/tiles) - A minimal desktop environment
  suitable for voice control.
* [Hiccup](https://git.sr.ht/~geb/hiccup) - Noise input for playing games.

## Support Me

[Thank you!](https://liberapay.com/geb)

## License

GPLv3 only, see [LICENSE](./LICENSE).

Copyright (c) 2022-2023 John Gebbie
