# Numen

Numen is voice control for computing without a keyboard. It works system-wide
on Linux and the speech recognition runs locally.

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

There shouldn't be any output but you should be able to type *hey* by saying
"hoof each yank".  You can also try transcribing a sentence after saying
"scribe", and terminate it by pressing Ctrl+c (a.k.a "troy cap").

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

## Keyboard Layouts

dotool will type gobbledygook if your environment has assigned it a different
keyboard layout than it's simulating keycodes for.  You can match them up by
setting the environment variables `DOTOOL_XKB_LAYOUT` and `DOTOOL_XKB_VARIANT`.

For example, if you use the French `fr` layout:

    DOTOOL_XKB_LAYOUT=fr numen

## Contact and Matrix Chat

You can send questions, thoughts or patches by composing an email to
[~geb/public-inbox@lists.sr.ht](https://lists.sr.ht/~geb/public-inbox).

You're also welcome to join our Matrix chat at
[#numen:matrix.org](https://matrix.to/#/#numen:matrix.org).

## See Also

* [Noggin](https://git.sr.ht/~geb/noggin) - face tracking I use for
  playing/developing games.
* [Tiles](https://git.sr.ht/~geb/tiles) - a minimal desktop environment
  suited to voice control.

## Support Me

[Thank you!](https://liberapay.com/geb)

## License

AGPLv3 only, see [LICENSE](./LICENSE).

Copyright (c) 2022-2023 John Gebbie
