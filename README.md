# Numen

Numen is voice control for handsfree computing, empowering people who have limited or no use of their hands.

## Installation From Source

The standard mode requires the [ydotool](https://github.com/ReimuNotMoe/ydotool) command, which can be installed with `sudo ./install-ydotool.sh`.
(Alternatively, you might be able to install `ydotool` from your package manager, but it needs to have been packaged in such a way it doesn't require root permissions to run.)

The speech recognition library and an english language model (about 40MB) are installed with `sudo ./install-vosk.sh`.

To install `numen` itself, you need `go` (>=1.15) and to run `sudo ./install-numen.sh`

## Getting Started

Once you've got a microphone, you can run it with: `numen`<br>
Say "hoof eve yank" to type "hey", or type a sentence like "scribe \<slight pause\> so this is a sentence blah blah blah".
You can terminate it by pressing <kbd>Ctrl</kbd>+<kbd>c</kbd> or saying "troll cap".

If nothing happened, you might need to specify the right microphone with the `--mic` option.
See `numen --list-mics` for what's available.

The default phrases are in the `/etc/numen/phrases/` directory, you could copy them to `~/.config/numen/phrases/` where you can edit them.
The most important files are `phrases/characters` with the alphabet and symbols, and `phrases/control` with the modifiers, backspace and friends.
Have a go in your text editor.

## Workflow

Voice control makes an efficient keyboard but a wack mouse.
At first I thought I'd need something like eye tracking, but now I just use keyboard based programs, which are thankfully the most productive kind of program.
The main two are [Neovim](https://neovim.io), my text editor, and [qutebrowser](https://qutebrowser.org), my web browser.
I also use a tiling window manager called [bspwm](https://github.com/baskerville/bspwm), so I don't need to arrange application windows with a mouse.

I'm planning to package for my full voice control set up.
