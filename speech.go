// speech MODEL_DIR PHRASE_FILE [PHRASE_FILE...]
package main

import (
	"bufio"
	"fmt"
	"git.sr.ht/~geb/vox"
	"github.com/go-audio/wav"
	"io"
	"os"
	"strings"
)

func fatal(a ...any) {
	fmt.Fprint(os.Stderr, "numen: ")
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

type Command struct {
	Phrase string
	Tags []string
	Action string
}

func get(cmds []Command, phrase string) (*Command, bool) {
	for i := range cmds {
		if cmds[i].Phrase == phrase {
			return &cmds[i], true
		}
	}
	return nil, false
}

func parse(paths []string, known func(string) bool, skip func([]string) bool) []Command {
	var commands []Command
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if s := strings.TrimSpace(sc.Text()); len(s) == 0 || []rune(s)[0] == '#' {
				continue
			}
			speech, action, found := strings.Cut(sc.Text(), ":")
			if len(action) > 0 {
				for []rune(action)[len([]rune(action))-1] == '\\' {
					if !sc.Scan() {
						fatal(f.Name() + ": unexpected end of file")
					}
					action = action[:len(action)-1] + "\n" + sc.Text()
				}
			}
			if !found {
				speech, _, _ = strings.Cut(speech, "#")
			}
			var tags []string
			var phrase string
			for _, field := range strings.Fields(speech) {
				if field[0] == '@' {
					if phrase != "" {
						fatal(f.Name() + ": all tags should be before the phrase: " + sc.Text())
					}
					tags = append(tags, field[1:])
				} else {
					if !known(field) {
						fatal(f.Name() + ": unknown word: " + field)
					}
					if phrase != "" {
						phrase += " "
					}
					phrase += field
				}
			}
			if !skip(tags) {
				if _, found := get(commands, phrase); found {
					fatal(f.Name() + ": phrase already defined: " + phrase)
				}
				commands = append(commands, Command{phrase, tags, action})
			}
		}
		if sc.Err() != nil {
			panic(sc.Err())
		}
	}
	return commands
}

func inspectAudio(r io.ReadSeeker) (int, int) {
	d := wav.NewDecoder(r)
	d.ReadInfo()
	if d.NumChans != 1 || d.WavAudioFormat != 1 {
		panic("audio must be the WAV format and mono")
	}
	return int(d.SampleRate), int(d.SampleBitDepth())
}

type EventType int
const (
	RapidOffEvent EventType = iota
	RapidOnEvent
	TranscribeEvent
)
type Event struct {
	Type EventType
	Content any
}

func printCommand(cmd *Command) {
	fmt.Println(cmd.Action)
	f := os.NewFile(4, "/dev/fd/4")
	if f != nil {
		fmt.Fprintln(f, cmd.Phrase)
	}
}

func printTranscripts(transRec *vox.Recognizer, cmd *Command) {
	for i, result := range transRec.FinalResults() {
		fmt.Printf("transcript%d:%s\n", i+1, result.Text)
	}
	printCommand(cmd)
}

func handle(cmds []Command, phrases []vox.PhraseResult, transRec *vox.Recognizer, audio []byte) []Event {
	cancel := 0
	CANCEL:
	for i := range phrases {
		c, _ := get(cmds, phrases[i].Text)
		for _, t := range c.Tags {
			if t == "transcribe" {
				break CANCEL
			}
			if t == "cancel" {
				cancel = i
			}
		}
	}
	phrases = phrases[cancel:]
	var events []Event
	for p := range phrases {
		c, _ := get(cmds, phrases[p].Text)
		transcribe := false
		for _, t := range c.Tags {
			switch t {
			case "transcribe":
				 transcribe = true
			case "rapidoff":
				events = append(events, Event{RapidOffEvent, nil})
			case "rapidon":
				events = append(events, Event{RapidOnEvent, nil})
			}
		}
		if transcribe {
			_, err := transRec.Accept(audio[phrases[p].End:])
			if err != nil {
				panic(err.Error())
			}
			if p == len(phrases)-1 {
				events = append(events, Event{TranscribeEvent, c})
			} else {
				printTranscripts(transRec, c)
			}
			break
		}
		printCommand(c)
	}
	return events
}

func main() {
	model, err := vox.NewModel(os.Args[1])
	if err != nil {
		fatal(err)
	}

	var commands []Command
	{
		known := func(s string) bool {
			return model.FindWord(s) != -1
		}
		handler := os.Getenv("NUMEN_HANDLER")
		skip := func(tags []string) bool {
			constrained := false
			for _, t := range tags {
				for _, h := range []string{"gadget", "kernel", "x11"} {
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
		commands = parse(os.Args[2:], known, skip)
		// For some reason Vosk starts outputting results close to "huh" if
		// you don't say anything for long enough.
		commands = append(commands, Command{"huh", nil, ""})
	}

	var cmdRec, transRec *vox.Recognizer
	{
		sampleRate, bitDepth := inspectAudio(os.Stdin)
		phrases := make([]string, len(commands))
		for i := range commands {
			phrases[i] = commands[i].Phrase
		}
		cmdRec, err = vox.NewRecognizer(model, sampleRate, bitDepth, phrases)
		if err != nil {
			panic(err.Error())
		}
		cmdRec.SetWords(true)
		cmdRec.SetKeyphrases(true)
		cmdRec.SetMaxAlternatives(3)

		transRec, err = vox.NewRecognizer(model, sampleRate, bitDepth, nil)
		if err != nil {
			panic(err.Error())
		}
		transRec.SetMaxAlternatives(10)
	}

	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 4096)

	var transcribing *Command
	rapid := false
	for {
		_, err :=  io.ReadFull(r, buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			panic(err.Error())
		}

		if transcribing == nil {
			finalized, err := cmdRec.Accept(buf)
			if err != nil {
				panic(err.Error())
			}
			if finalized || (rapid && cmdRec.Results()[0].Text != "") {
				for _, result := range cmdRec.FinalResults() {
					phrases := result.Phrases
					ok := result.Valid
					if !ok {
						for p := range phrases {
							c, _ := get(commands, phrases[p].Text)
							for _, t := range c.Tags {
								if t == "transcribe" {
									ok = true
									break
								}
							}
						}
					}
					if ok {
						events := handle(commands, phrases, transRec, cmdRec.Audio)
						for _, e := range events {
							switch e.Type {
							case RapidOffEvent:
								rapid = false
							case RapidOnEvent:
								rapid = true
							case TranscribeEvent:
								transcribing = e.Content.(*Command)
							}
						}
						break
					}
				}
			}
		} else {
			finalized, err := transRec.Accept(buf)
			if err != nil {
				panic(err.Error())
			}
			if finalized {
				printTranscripts(transRec, transcribing)
				transcribing = nil
			}
		}
	}

	// We don't handle any final bit of audio.
}
