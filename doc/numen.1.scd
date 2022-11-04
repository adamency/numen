numen(1)

# NAME

*numen* - voice control for handsfree computing

# SYNOPSIS

*numen* [*--kernel*|*--x11*] [*--mic=*_NAME_] [_FILE_...]++
*numen* *--help*

# DESCRIPTION

*numen* reads phrases and corresponding actions from one or more files and
performs the actions when you say the phrases.

When no files are passed as arguments, *numen* reads the files ending
with .phrases in *$XDG_CONFIG_HOME/numen/phrases/* else */etc/numen/phrases/*.
(*$XDG_CONFIG_HOME* is *~/.config* unless set otherwise.)

The actions can be performed for different environments.  By default, a
virtual device is simulated at a low enough level to work in any graphical
environment or virtual console.

# OPTIONS

*-h*, *--help*
	Print help and exit.

*--kernel*
	The default mode, perform the actions at a low enough level to work in
	any graphical environment or virtual console.  This requires permission
	to */dev/uinput* which the udev rule installed with numen grants to
	users in group *input*.  If you just installed numen, you can make the
	rule effective without rebooting by running: *sudo udevadm trigger*.

*--list-mics*
	List audio devices and exit.

*--mic=*_NAME_
	Select the audio device.

*--x11*
	Perform the actions for an X11 graphical environment.  This is more
	limited than *--kernel* but requires less permissions.

# CONFIGURATION

Each line of the files goes:

\[*@*_TAG_...] _PHRASE_:_ACTIONS_

That is, there can be zero or more "tags" each starting with *@*, then the
phrase, then a *:* and the actions.  Empty lines and lines starting with *#*
are ignored.  The *:* can be omitted if there are no actions.

Currently, the phrase must be a single word.  The phrase must be known by
the speech recognition or you'll be informed by an error.

Actions can come directly after other actions that don't use the rest of
the line.  Multiple actions can be spread over multiple lines by ending each
with a backslash.  The supported actions are:

*press* _CHORD_ [_CHORD_...]
	Simulate each _CHORD_, for example: *press ctrl+c Up*++
The key names are the X11 keysyms, which you can find with the *xev* command.
	Not all keys are supported yet in all modes.

*type* _TEXT_
	Simulate typing the rest of the line, for example:++
*type Hello, World!*

*mod* *shift*/*ctrl*/*alt*/*super*
	Enable a modifier for the next press.

*mod* *clear*
	Clear all modifiers.

*capson*/*capsoff*
	Enable/disable Caps Lock.

*mouseto* _X_ _Y_
	Teleport the cursor to _X_ _Y_, where _X_ and _Y_ are percentages
	between 0.0 and 1.0.

*mousemove* _X_ _Y_
	Move the cursor _X_ _Y_ pixels relative to its current position.

*click* *left*/*middle*/*right*
	Simulate a mouse click.

*scrollup*/*scrolldown*
	Simulate scrolling a mouse wheel.

*stick*
	Make presses not keyup until *unstick*, like holding down a key at
	a time.  If already sticking, keyup but keep sticking.

*unstick*
	Keyup and stop sticking.

*repeat* _NUMBER_
	Repeat the previous *press*/*pen* the _NUMBER_ of times.

*pen* _COMMAND_
	Execute the rest of the line in the shell and type any stdout.
	An example that types the date: *pen date +%D*

*eval* _COMMAND_
	Reinterpret the rest of the line as the stdout it produces when shell
	evaluated.  An example that backspaces the previously typed text:++
*eval awk -v n=${#NUMEN_TEXT} 'BEGIN {for(; n--; n) print "press BackSpace"}'*

*set* _ENVIRONMENT_VARIABLE_ _COMMAND_
	Set _ENVIRONMENT_VARIABLE_ to the stdout the rest of the line produces
	when shell evaluated.  An example that tweaks *numen*:++
*set NUMEN_TYPE_DELAY echo 20*

*handler* *kernel*/*x11*
	Change between the *--kernel* and *--x11* modes.

Tags have phrases act specially.  Unsupported tags are silently ignored.
The supported tags are:

*@instant*
	If the phrase seems to have been said, instead of waiting till the
	sentence is finished and finalized, perform the sentence so far.

*@cancel*
	Cancel the sentence.  Note without *@instant* preceding words are still
	performed, but with *@instant* proceeding words are still performed.

*@transcribe*
	Instead of performing phrases, listen till the end of the sentence,
	then set the environment variable *NUMEN_TRANSCRIPT* to the speech
	recognition result and perform the action.  An example that would
	transcribe the rest of the sentence uppercase:++
*@instant @transcribe scream:pen echo $NUMEN_TRANSCRIPT | tr '[:lower:]' '[:upper:]'*

*@rapidon*
	Act as if all phrases are tagged with *@instant* until a phrase
	tagged *@rapidoff*.  Useful to shorten the response time in *sticky*
	mode, but reduces recognition accuracy.

*@kernel* *@x11*
	Only listen for the phrase if performing the actions using one of
	the tagged modes.

# ENVIRONMENT VARIABLES

*numen* can be tweaked by some environment variables, which can even be set
by the *set* action:

- *NUMEN_SHELL* sets the shell used for *pen* and *eval* actions.
- *NUMEN_TYPE_DELAY* sets the keystroke delay of the *type* and *pen* actions.
- *NUMEN_KEY_DELAY* sets the keystroke delay of the *press* action.

When *numen* executes a shell command, it provides some environment variables
exposing *numen*'s state:

- *NUMEN_HANDLER* is *kernel* or *x11* depending on which mode the actions
  are being performed.
- *NUMEN_TEXT* is the last *type*/*pen* text.
- *NUMEN_TRANSCRIPT* is the literal sentence from a *@transcribe* phrase.
- *NUMEN_SUPER*, *NUMEN_CTRL*, *NUMEN_ALT* and *NUMEN_SHIFT* are either *0*
  or *1* describing each modifier's state.

# DATA FILES

The file *$XDG_DATA_HOME/numen/transcripts* is overwritten to the latest
transcription results.

The directory *$XDG_DATA_HOME/numen/mods/* has files of the state of each
modifier, containing either *0* or *1*.

(*$XDG_DATA_HOME* is *~/.local/share/* unless set otherwise.)

# EXAMPLES

Run *numen* using the default phrases in */etc/numen/phrases/*:

	*numen*

You can try saying "hoof eve yank" and "scribe, so this is a sentence"
into your microphone, and terminate it by pressing *Ctrl*+*c* or saying
"troll cap".

# AUTHOR

John Gebbie
