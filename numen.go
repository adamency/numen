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
	DefaultModel        = "/usr/share/vosk-models/small-en-us"
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
--models=PATHS   Specify the speech recognition models.
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

type Recognition struct {
	Models   []*vosk.VoskModel
	Actions map[string]Action

	CmdRec   *vox.Recognizer
	TranRecs []*vox.Recognizer
	TranRec  *vox.Recognizer
	NoiseRec *NoiseRecognizer
	NoiseBuffer *bytes.Buffer
}

func (r *Recognition) LoadPhrases(files []string) error {
	var err error
	r.Actions, err = parseFiles(files, opts.Handler, r.Models[0])
	if err != nil {
		return err
	}
	if r.CmdRec != nil {
		r.CmdRec.SetGrm(getPhrases(r.Actions))
	}

	if blow, hiss, shush := haveNoises(r.Actions); blow || hiss || shush {
		r.NoiseBuffer = bytes.NewBuffer([]byte(wavHeader))
		r.NoiseRec = NewNoiseRecognizer(r.NoiseBuffer, blow, hiss, shush)
	} else {
		r.NoiseBuffer = nil
		r.NoiseRec = nil
	}
	return nil
}

func (r *Recognition) Free() {
	for i := range r.Models {
		r.Models[i].Free()
	}
	if r.CmdRec != nil {
		r.CmdRec.Free()
	}
	for i := range r.TranRecs {
		r.TranRecs[i].Free()
	}
}

func loadModels() []*vosk.VoskModel {
	if opts.Models == "" {
		// backwards compatibility
		opts.Models = os.Getenv("NUMEN_MODEL")
	}

	if opts.Models == "" {
		if _, err := os.Stat(DefaultModel); errors.Is(err, os.ErrNotExist) {
			fatal("The default model doesn't exist: " + DefaultModel + `
so specify --models or install the default model package: ` + DefaultModelPackage)
		}
		opts.Models = DefaultModel
	}

	fields := strings.Fields(opts.Models)
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Models: %q\n", fields)
	}
	models := make([]*vosk.VoskModel, len(fields))
	for i := range fields {
		var err error
		models[i], err = vox.NewModel(fields[i])
		if err != nil {
			fatal(err)
		}
	}
	return models
}

func makeRecognizers(models []*vosk.VoskModel, phrases []string) (cmdRec *vox.Recognizer, tranRecs []*vox.Recognizer) {
	sampleRate, bitDepth := 16000, 16
	var err error
	cmdRec, err = vox.NewRecognizer(models[0], sampleRate, bitDepth, phrases)
	if err != nil {
		panic(err)
	}
	cmdRec.SetWords(true)
	cmdRec.SetKeyphrases(true)
	cmdRec.SetMaxAlternatives(3)

	tranRecs = make([]*vox.Recognizer, len(models))
	for i := range tranRecs {
		var err error
		tranRecs[i], err = vox.NewRecognizer(models[i], sampleRate, bitDepth, nil)
		if err != nil {
			panic(err)
		}
		tranRecs[i].SetMaxAlternatives(10)
	}
	return
}

var opts struct {
	Audio     string
	AudioLog  *os.File
	Files     []string
	Handler   string
	Mic       string
	Models    string
	PhraseLog *os.File
	Verbose   bool
}

func main() {
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

		o.Func("models", func(s string) error {
			opts.Models = s
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

	rec := &Recognition{}
	defer rec.Free()

	rec.Models = loadModels()

	if err := rec.LoadPhrases(opts.Files); err != nil {
		fatal(err)
	}

	rec.CmdRec, rec.TranRecs = makeRecognizers(rec.Models, getPhrases(rec.Actions))
	rec.TranRec = rec.TranRecs[0]

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

	var handler *Handler
	{
		if opts.Handler == "gadget" {
			h := Handler(NewGadgetHandler(rec))
			handler = &h
		} else if opts.Handler == "uinput" {
			h := Handler(NewUinputHandler(rec))
			handler = &h
		} else if opts.Handler == "x11" {
			h := Handler(NewX11Handler(rec))
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

		if len(rec.Actions) == 0 {
			continue
		}

		if transcribing == "" {
			var finalized bool

			if rec.NoiseRec != nil {
				rec.NoiseBuffer.Write(chunk)
				rec.NoiseRec.Proceed(len(chunk) / 2)
				if rec.NoiseRec.Noise != rec.NoiseRec.PrevNoise {
					if s := noiseEndString(rec.NoiseRec.PrevNoise); s != "" {
						handle(handler, rec.Actions[s].Text)
						writeLine(opts.PhraseLog, s)
					}
					if s := noiseBeginString(rec.NoiseRec.Noise); s != "" {
						handle(handler, rec.Actions[s].Text)
						writeLine(opts.PhraseLog, s)
						finalized = true
					}
				}
				if !finalized && rec.NoiseRec.Noise != NoiseNone {
					continue
				}
			}

			if !finalized {
				var err error
				finalized, err = rec.CmdRec.Accept(chunk)
				if err != nil {
					panic(err)
				}
			}
			if finalized || ((*handler).Sticky() && rec.CmdRec.Results()[0].Text != "") {
				var result vox.Result
				var valid bool
				for _, result = range rec.CmdRec.FinalResults() {
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
						for _, t := range rec.Actions[phrase.Text].Tags {
							if t == "transcribe" {
								valid = true
								break PHRASE
							}
						}
					}

					if valid {
						transcribing = do(rec.CmdRec, rec.TranRec, handler, sentence, rec.Actions, rec.CmdRec.Audio, opts.PhraseLog)
						if transcribing == "" {
							handle(handler, rec.Actions["<complete>"].Text)
						}
						break
					}
				}

				if !valid {
					if a, ok := rec.Actions["<unknown>"]; ok {
						writeLine(opts.PhraseLog, "<unknown>")
						handle(handler, a.Text)
					}
				}
			}
		} else {
			finalized, err := rec.TranRec.Accept(chunk)
			if err != nil {
				panic(err)
			}
			if finalized {
				handleTranscribe(handler, rec.TranRec.FinalResults(), rec.Actions[transcribing])
				writeLine(opts.PhraseLog, transcribing)
				handle(handler, rec.Actions["<complete>"].Text)
				transcribing = ""
			}
		}
	}
	// TODO Handle any final bit of audio.
}
