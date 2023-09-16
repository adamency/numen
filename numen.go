package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"git.sr.ht/~geb/numen/vox"
	"git.sr.ht/~geb/opt"
	vosk "github.com/alphacep/vosk-api/go"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	Version             string
	DefaultModelPackage = "vosk-model-small-en-us"
	DefaultModelPaths   = "/usr/share/vosk-models/small-en-us"
	DefaultPhrasesDir   = "/etc/numen/phrases"
)

func usage() {
	fmt.Println(`Usage: numen [FILE...]

numen reads phrases and actions from one or more files, and performs the
actions when you say their phrases.

--audio=FILE     Specify an audio file to use instead of the microphone.
--audiolog=FILE  Write the audio to FILE while it's recorded.
--gadget         Use the gadget handler to perform the actions over USB.
--uinput         Use the uinput handler to perform the actions. (default)
--list-mics      List audio devices and exit. (same as arecord -L)
--mic=NAME       Specify the audio device.
--phraselog=FILE Write phrases to FILE when they are performed.
--verbose        Show what is being used.
--version        Print the version and exit.
--x11            Use the X11 handler to perform the actions.`)
}

func fatal(a ...any) {
	fmt.Fprintln(os.Stderr, "numen:", fmt.Sprint(a...))
	os.Exit(1)
}
func warn(a ...any) {
	fmt.Fprintln(os.Stderr, "numen: WARNING:", fmt.Sprint(a...))
}

func pipeBeingRead(path string) bool {
	opened := make(chan bool)
	go func() {
		f, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
		opened <- err == nil
		if err == nil {
			f.Close()
		}
	}()
	select {
	case ok := <-opened:
		return ok
	case <-time.After(time.Millisecond):
		return false
	}
}

func writeLine(f *os.File, s string) {
	if f != nil {
		_, err := io.WriteString(f, s+"\n")
		if err != nil {
			warn(err)
		}
	}
}

func init() {
	p := os.Getenv("NUMEN_STATE_DIR")
	if p == "" {
		p = os.Getenv("XDG_STATE_HOME")
		if p == "" {
			p = os.Getenv("HOME")
			if p == "" {
				warn("not $NUMEN_STATE_DIR nor $XDG_STATE_HOME nor $HOME is defined")
				return
			}
			p += "/.local/state"
		}
	}
	p += "/numen"
	err := os.MkdirAll(p, 0o700)
	if err != nil {
		fatal(err)
	}
	os.Setenv("NUMEN_STATE_DIR", p)
}
func writeStateFile(name string, data []byte) {
	err := os.WriteFile(os.Getenv("NUMEN_STATE_DIR")+"/"+name, data, 0o600)
	if err != nil {
		warn(err)
	}
}

type Action struct {
	Tags []string
	Text string
}

func knownSpecialPhrase(phrase string) bool {
	switch phrase {
	case "<complete>":
		return true
	case "<blow-begin>", "<blow-end>":
		return true
	case "<hiss-begin>", "<hiss-end>":
		return true
	case "<shush-begin>", "<shush-end>":
		return true
	case "<unknown>":
		return true
	}
	return false
}
func knownTag(tag string) bool {
	switch tag {
	case "cancel", "gadget", "uinput", "transcribe", "x11":
		return true
	}
	return false
}
func skipPhrase(tags []string, handler string) bool {
	constrained := false
	for _, t := range tags {
		for _, h := range []string{"gadget", "uinput", "x11"} {
			if t == h && h == handler {
				return false
			}
			if t == h {
				constrained = true
			}
		}
	}
	return constrained
}

func parseFiles(paths []string, handler string, model *vosk.VoskModel) (map[string]Action, error) {
	actions := make(map[string]Action)
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return actions, err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if s := strings.TrimSpace(sc.Text()); len(s) == 0 || []rune(s)[0] == '#' {
				continue
			}
			speech, action, found := strings.Cut(sc.Text(), ":")
			if !found {
				warn(f.Name() + ": invalid phrase definition: " + sc.Text())
				continue
			}
			if len(action) > 0 {
				for []rune(action)[len([]rune(action))-1] == '\\' {
					if !sc.Scan() {
						warn(f.Name() + ": unexpected end of file")
						break
					}
					action = action[:len(action)-1] + "\n" + sc.Text()
				}
			}
			var tags []string
			var phrase string
			for _, field := range strings.Fields(speech) {
				if field[0] == '@' {
					if phrase != "" {
						warn(f.Name() + ": all tags should be before the phrase: " + speech)
						phrase = ""
						break
					}
					if knownTag(field[1:]) {
						tags = append(tags, field[1:])
					} else {
						warn(f.Name() + ": unknown tag: " + field)
					}
				} else if knownSpecialPhrase(field) {
					if phrase != "" {
						warn(f.Name() + ": special phrases can't be mixed with words: " + speech)
						phrase = ""
						break
					}
					phrase += field
				} else {
					if phrase != "" && phrase[0] == '<' {
						warn(f.Name() + ": special phrases can't be mixed with words: " + speech)
						phrase = ""
						break
					}
					if model.FindWord(field) == -1 {
						warn(f.Name() + ": phrase with unknown word: " + speech)
						phrase = ""
						break
					}
					if phrase != "" {
						phrase += " "
					}
					phrase += field
				}
			}
			if phrase != "" && !skipPhrase(tags, handler) {
				if _, ok := actions[phrase]; ok {
					warn(f.Name() + ": phrase redefined: " + phrase)
				}
				actions[phrase] = Action{tags, action}
			}
		}
		if sc.Err() != nil {
			panic(sc.Err())
		}
	}
	return actions, nil
}

func getPhrases(actions map[string]Action) []string {
	phrases := make([]string, 0, len(actions))
	for p := range actions {
		if p == "<unknown>" {
			phrases = append(phrases, "[unk]")
		} else if p[0] != '<' {
			phrases = append(phrases, p)
		}
	}
	return phrases
}

func haveNoises(actions map[string]Action) (bool, bool, bool) {
	var blow, hiss, shush bool
	for p := range actions {
		if strings.HasPrefix(p, "<blow-") {
			blow = true
		} else if strings.HasPrefix(p, "<hiss-") {
			hiss = true
		} else if strings.HasPrefix(p, "<shush-") {
			shush = true
		}
	}
	return blow, hiss, shush
}

func handleTranscribe(h *Handler, results []vox.Result, action Action) {
	var b bytes.Buffer
	for _, r := range results {
		b.WriteString(r.Text + "\n")
	}
	writeStateFile("transcripts", b.Bytes())
	handle(h, action.Text)
}

func do(cmdRec, transRec *vox.Recognizer, handler *Handler, sentence []vox.PhraseResult, actions map[string]Action, audio []byte, phraseLog *os.File) string {
	cancel := 0
CANCEL:
	for i := range sentence {
		act, _ := actions[sentence[i].Text]
		for _, tag := range act.Tags {
			if tag == "transcribe" {
				break CANCEL
			}
			if tag == "cancel" {
				cancel = i
			}
		}
	}
	sentence = sentence[cancel:]

	for i := range sentence {
		phrase := sentence[i].Text
		act, _ := actions[phrase]
		transcribe := false
		for _, tag := range act.Tags {
			if tag == "transcribe" {
				transcribe = true
			}
		}
		if transcribe {
			_, err := transRec.Accept(audio[sentence[i].End:])
			if err != nil {
				panic(err)
			}
			if i == len(sentence)-1 {
				return phrase
			}
			handleTranscribe(handler, transRec.FinalResults(), act)
			writeLine(phraseLog, phrase)
			return ""
		}
		handle(handler, act.Text)
		writeLine(phraseLog, phrase)
		writeStateFile("phrase", []byte(phrase))
	}

	// Carrying over helps especially when there is no required pause.
	trailing := cmdRec.Audio[sentence[len(sentence)-1].End:]
	_, err := cmdRec.Accept(trailing)
	if err != nil {
		panic(err)
	}
	return ""
}

func main() {
	var opts struct {
		AudioLog  *os.File
		Files     []string
		Handler   string
		Mic       string
		PhraseLog *os.File
		Verbose   bool
	}
	opts.Handler = "uinput"
	audio := &Audio{}
	{
		o := opt.NewOptionSet()

		o.Func("audio", func(s string) error {
			audio.Filename = s
			return nil
		})

		o.Func("audiolog", func(s string) error {
			var err error
			opts.AudioLog, err = os.Create(s)
			if err != nil {
				fatal(err)
			}
			return nil
		})

		o.FlagFunc("gadget", func() error {
			opts.Handler = "gadget"
			return nil
		})

		o.FlagFunc("h", func() error {
			usage()
			os.Exit(0)
			panic("unreachable")
		})
		o.Alias("h", "help")

		o.FlagFunc("list-mics", func() error {
			cmd := exec.Command("arecord", "-L")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fatal(err)
			}
			os.Exit(0)
			panic("unreachable")
		})

		o.Func("mic", func(s string) error {
			opts.Mic = s
			return nil
		})

		o.Func("phraselog", func(s string) error {
			var err error
			opts.PhraseLog, err = os.Create(s)
			if err != nil {
				fatal(err)
			}
			return nil
		})

		o.FlagFunc("uinput", func() error {
			opts.Handler = "uinput"
			return nil
		})

		o.BoolFunc("verbose", func(b bool) error {
			opts.Verbose = b
			return nil
		})

		o.FlagFunc("version", func() error {
			fmt.Println(Version)
			os.Exit(0)
			panic("unreachable")
		})

		o.FlagFunc("x11", func() error {
			opts.Handler = "x11"
			return nil
		})

		err := o.Parse(true, os.Args[1:])
		if err != nil {
			fatal(err)
		}
		if len(o.Args()) > 0 {
			opts.Files = o.Args()
		} else {
			p, err := os.UserConfigDir()
			if err == nil {
				opts.Files, err = filepath.Glob(p + "/numen/phrases/*.phrases")
				if err != nil {
					panic(err)
				}
			}
			if opts.Files == nil {
				opts.Files, err = filepath.Glob(DefaultPhrasesDir + "/*.phrases")
				if err != nil {
					panic(err)
				}
				if opts.Files == nil {
					fatal("the default phrase files are missing?!")
				}
			}
		}
	}
	if opts.AudioLog != nil {
		defer opts.AudioLog.Close()
	}
	if opts.PhraseLog != nil {
		defer opts.PhraseLog.Close()
	}
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Files: %q\n", opts.Files)
	}
	writeStateFile("handler", []byte(opts.Handler))

	var model *vosk.VoskModel
	{
		m := os.Getenv("NUMEN_MODEL")
		if m == "" {
			for _, p := range strings.Fields(DefaultModelPaths) {
				if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
					m = p
					break
				}
			}
		}
		if m == "" {
			fatal("you need to install the " + DefaultModelPackage + " package or set $NUMEN_MODEL")
		}
		if opts.Verbose {
			fmt.Fprintln(os.Stderr, "Model: "+m)
		}

		var err error
		model, err = vox.NewModel(m)
		if err != nil {
			fatal(err)
		}
	}
	defer model.Free()

	actions, err := parseFiles(opts.Files, opts.Handler, model)
	if err != nil {
		fatal(err)
	}

	var cmdRec, transRec *vox.Recognizer
	{
		sampleRate, bitDepth := 16000, 16
		var err error
		cmdRec, err = vox.NewRecognizer(model, sampleRate, bitDepth, getPhrases(actions))
		if err != nil {
			panic(err)
		}
		cmdRec.SetWords(true)
		cmdRec.SetKeyphrases(true)
		cmdRec.SetMaxAlternatives(3)

		transRec, err = vox.NewRecognizer(model, sampleRate, bitDepth, nil)
		if err != nil {
			panic(err)
		}
		transRec.SetMaxAlternatives(10)
	}
	defer cmdRec.Free()
	defer transRec.Free()

	if audio.Filename == "" {
		audio.SetDevice(opts.Mic)
		if opts.Verbose {
			fmt.Fprintln(os.Stderr, "Microphone: "+audio.Device)
		}
	}
	if err := audio.Start(); err != nil {
		fatal(err)
	}
	defer audio.Close()

	var noiseRec *NoiseRecognizer
	var noiseBuffer *bytes.Buffer
	if blow, hiss, shush := haveNoises(actions); blow || hiss || shush {
		noiseBuffer = new(bytes.Buffer)
		noiseRec = NewNoiseRecognizer(noiseBuffer, blow, hiss, shush)
	}

	var handler *Handler
	{
		load := func(files []string) {
			acts, err := parseFiles(files, opts.Handler, model)
			if err != nil {
				warn(err)
				return
			}
			actions = acts
			cmdRec.SetGrm(getPhrases(actions))

			if blow, hiss, shush := haveNoises(actions); blow || hiss || shush {
				noiseBuffer = bytes.NewBuffer([]byte(wavHeader))
				noiseRec = NewNoiseRecognizer(noiseBuffer, blow, hiss, shush)
			} else {
				noiseBuffer = nil
				noiseRec = nil
			}
		}
		if opts.Handler == "gadget" {
			h := Handler(NewGadgetHandler(load))
			handler = &h
		} else if opts.Handler == "uinput" {
			h := Handler(NewUinputHandler(load))
			handler = &h
		} else if opts.Handler == "x11" {
			h := Handler(NewX11Handler(load))
			handler = &h
		} else {
			panic("unreachable")
		}
		defer func() { (*handler).Close() }()
	}

	pipe := make(chan func())
	{
		p := os.Getenv("NUMEN_PIPE")
		if p == "" {
			p = "/tmp/numen-pipe"
		}
		if opts.Verbose {
			fmt.Fprintln(os.Stderr, "Pipe: "+p)
		}

		if pipeBeingRead(p) {
			fatal("another instance is already reading the pipe: " + p)
		}

		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			fatal(err)
		}
		if err := syscall.Mkfifo(p, 0o600); err != nil {
			panic(err)
		}
		defer os.Remove(p)
		f, err := os.OpenFile(p, os.O_RDWR, os.ModeNamedPipe)
		if err != nil {
			panic(err)
		}

		go func() {
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				pipe <- func() { handle(handler, sc.Text()) }
			}
			if sc.Err() != nil {
				warn(sc.Err())
			}
		}()
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	retry := false
	transcribing := ""
	for {
		select {
		case <-terminate:
			return
		case f := <-pipe:
			f()
		default:
		}
		chunk := make([]byte, 4096)
		_, err := io.ReadFull(audio.Reader(), chunk)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				if audio.Filename == "" && retry {
					_ = audio.Close()
					if err := audio.Start(); err != nil {
						warn(err)
					}
					continue
				}
				return
			}
			panic(err)
		}
		retry = true
		if opts.AudioLog != nil {
			if _, err := opts.AudioLog.Write(chunk); err != nil {
				warn(err)
			}
		}

		if len(actions) == 0 {
			continue
		}

		if transcribing == "" {
			var finalized bool

			if noiseRec != nil {
				noiseBuffer.Write(chunk)
				noiseRec.Proceed(len(chunk) / 2)
				if noiseRec.Noise != noiseRec.PrevNoise {
					if s := noiseEndString(noiseRec.PrevNoise); s != "" {
						handle(handler, actions[s].Text)
						writeLine(opts.PhraseLog, s)
					}
					if s := noiseBeginString(noiseRec.Noise); s != "" {
						handle(handler, actions[s].Text)
						writeLine(opts.PhraseLog, s)
						finalized = true
					}
				}
				if !finalized && noiseRec.Noise != NoiseNone {
					continue
				}
			}

			if !finalized {
				var err error
				finalized, err = cmdRec.Accept(chunk)
				if err != nil {
					panic(err)
				}
			}
			if finalized || ((*handler).Sticky() && cmdRec.Results()[0].Text != "") {
				var result vox.Result
				var valid bool
				for _, result = range cmdRec.FinalResults() {
					if result.Text == "" {
						continue
					}
					sentence := result.Phrases
					valid = result.Valid
PHRASE:
					for _, phrase := range sentence {
						if phrase.Text == "[unk]" {
							valid = false
							break
						}
						for _, t := range actions[phrase.Text].Tags {
							if t == "transcribe" {
								valid = true
								break PHRASE
							}
						}
					}

					if valid {
						transcribing = do(cmdRec, transRec, handler, sentence, actions, cmdRec.Audio, opts.PhraseLog)
						if transcribing == "" {
							handle(handler, actions["<complete>"].Text)
						}
						break
					}
				}

				if !valid {
					if a, ok := actions["<unknown>"]; ok {
						writeLine(opts.PhraseLog, "<unknown>")
						handle(handler, a.Text)
					}
				}
			}
		} else {
			finalized, err := transRec.Accept(chunk)
			if err != nil {
				panic(err)
			}
			if finalized {
				handleTranscribe(handler, transRec.FinalResults(), actions[transcribing])
				writeLine(opts.PhraseLog, transcribing)
				handle(handler, actions["<complete>"].Text)
				transcribing = ""
			}
		}
	}
	// TODO Handle any final bit of audio.
}
