# Numen

Numen is voice control for desktop computing without a keyboard or mouse.
It works system-wide on your Linux machine.

A short demonstration can be found on: [https://numen.johngebbie.com](https://numen.johngebbie.com)

## Installation From Source

The standard mode requires the [ydotool](https://github.com/ReimuNotMoe/ydotool) command, which can be installed with `sudo ./install-ydotool.sh`.
(Alternatively, you might be able to install `ydotool` from your package manager, but it needs to have been packaged in such a way it doesn't require root permissions to run.)

The [speech recognition library](https://alphacephei.com/vosk) and an english language model (about 40MB) can be installed with `sudo ./install-vosk.sh`.

Finally, `numen` itself requires `go` (>=1.15) and can be installed with `sudo ./install-numen.sh`.

## Getting Started

Once you've got a microphone, you can run it with: `numen`<br>
You should be able to type "hey" by saying "hoof eve yank", and transcribe a sentence after saying "scribe" (you need to leave a slight pause after "scribe" for now).
You can terminate it by pressing <kbd>Ctrl</kbd>+<kbd>c</kbd> or saying "troll cap".

If nothing happened, you might need to specify the right microphone with the `--mic` option.
See `numen --list-mics` for what's available.

The default phrases are in the `/etc/numen/phrases/` directory, and you can copy them to `~/.config/numen/phrases/` where you can edit them.
The most important files are `phrases/characters` with the alphabet and symbols, and `phrases/control` with the modifiers, backspace and friends.
Have a go in your text editor.

## Going Further

Voice control makes an efficient keyboard but a wack mouse.
At first I thought I'd need something like eye tracking, but now I just use keyboard based programs, which are thankfully the most productive kind of program.
The main two are [Neovim](https://neovim.io), my text editor, and [qutebrowser](https://qutebrowser.org), my web browser.

I've also made a desktop environment that works well with voice control, called [tiles](https://git.sr.ht/~geb/tiles).

## Contact

You can ask for help or send patches by composing an email to [~geb/public-inbox@lists.sr.ht](https://lists.sr.ht/~geb/public-inbox).
You're also welcome to join our Matrix chat at [#numen:matrix.org](https://matrix.to/#/#numen:matrix.org).
