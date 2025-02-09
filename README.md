# Numen

Numen is voice control for handsfree computing, letting you type efficiently
by saying syllables and literal words. It works system-wide on Linux and
the speech recognition runs locally.

There's a short demonstration on: [numenvoice.org](https://numenvoice.org)

## Install From Packages

Packages of Numen are available on:

- [Alpine](https://pkgs.alpinelinux.org/packages?name=numen)
- [Arch (AUR)](https://aur.archlinux.org/packages?SeB=N&K=numen)
- [Nix (anpandey's flake)](https://github.com/anpandey/numen-nix)

and potentially other platforms.

## Install From Source

`go` (aka `golang`) and `scdoc` are required.

A binary of the [speech recognition library](https://alphacephei.com/vosk)
and an English model (about 40MB) can be installed with:

    ./get-vosk.sh && sudo ./get-vosk.sh install
    ./get-model.sh && sudo ./get-model.sh install

The [dotool](https://sr.ht/~geb/dotool) command which simulates the input,
can be installed with:

    ./get-dotool.sh && sudo ./get-dotool.sh install

Finally, Numen itself can be installed with:

    ./build.sh && sudo ./build.sh install

## Permission and Keyboard Layouts

`dotool` requires permission to `/dev/uinput` to create the virtual input
devices, and a udev rule grants this to users in group input.

You can try:

    echo type hello | dotool

and if need be, you could add your user to group input with:

    sudo groupadd -f input
    sudo usermod -a -G input $USER

and re-login and trigger the udev rule or just reboot.

If it types something other than *hello*, see about keyboard layouts in the
[manpage](doc/numen.1.scd).

## Getting Started

Once you've got a microphone, you can run it with:

    numen

There shouldn't be any output, but you can try typing *hey* by saying "hoof
each yank", and try transcribing a sentence after saying "scribe". Terminate
it by pressing Ctrl+c (aka "troy cap").

If nothing happened, check it's using the right audio device with:

    timeout 5 numen --verbose --audiolog=me.wav
    aplay me.wav

and specify a `--mic` from `--list-mics` if not.

Now you're ready to have a go in your text editor! The default phrases are
in the `/etc/numen/phrases` directory.

## Going Further

I use Numen and the default phrases for all my computing, with
keyboard-based programs like [Neovim](https://neovim.io) and
[qutebrowser](https://qutebrowser.org). I also use a minimal desktop
environment I made, called [Tiles](https://git.sr.ht/~geb/tiles), that
doesn't require a pointer device for window management, file picking, etc.

If you'd like to tweak the phrases, copy the default phrases to
`~/.config/numen/phrases` and edit them there. The [manpage](doc/numen.1.scd)
covers configuration.

## Mailing List and Matrix Chat

You can send questions or patches by composing an email to
[~geb/numen@lists.sr.ht](https://lists.sr.ht/~geb/numen).

You're also welcome to join the Matrix chat at
[#numen:matrix.org](https://matrix.to/#/#numen:matrix.org).

## See Also

* [awesome-numen](https://git.sr.ht/~geb/awesome-numen) - a list of Numen
  configs and resources
* [sprec](https://git.sr.ht/~geb/sprec) - a speech recognition command
  (if you're just looking for speech to text)
* [Tiles](https://git.sr.ht/~geb/tiles) - a minimal desktop environment
  suited to voice control
* [Noggin](https://git.sr.ht/~geb/noggin) - experimental face tracking I
  made for playing games

## Support My Work 👀

[Thank you!](https://liberapay.com/geb)

## License

AGPLv3 only, see [LICENSE](./LICENSE).

Copyright (c) 2022-2024 John Gebbie

## Extra Responsive Hack

You can append these lines to your model's *conf/model.conf* to make things
extra responsive:

    --endpoint.rule2.min-trailing-silence=0.25
    --endpoint.rule3.min-trailing-silence=0.25
    --endpoint.rule4.min-trailing-silence=0.3

The default model is */usr/share/vosk-models/small-en-us*, but you can edit
a copy instead and specify it with the *NUMEN_MODEL* environment variable. We
should be able to implement better with the next Vosk release.
