// speech MODEL_DIR PHRASE_FILE [PHRASE_FILE...]
package main

// TODO Confidence threshold for @instant words.

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
			phrase, action, found := strings.Cut(sc.Text(), ":")
			if len(action) > 0 {
				for []rune(action)[len([]rune(action))-1] == '\\' {
					if !sc.Scan() {
						fatal(f.Name() + ": unexpected end of file")
					}
					action = action[:len(action)-1] + "\n" + sc.Text()
				}
			}
			if !found {
				phrase, _, _ = strings.Cut(phrase, "#")
			}
			fields := strings.Fields(phrase)
			if len(fields) == 0 {
				fatal(f.Name() + ": line missing phrase: " + sc.Text())
			}
			phrase = fields[len(fields)-1]
			tags := fields[:len(fields)-1]
			for i := range tags {
				if tags[i][0] != '@' {
					fatal(f.Name() + ": currently only one-word phrases are supported: " + sc.Text())
				}
				tags[i] = tags[i][1:]
			}
			if !known(phrase) {
				fatal(f.Name() + ": phrase not in the vocabulary: " + phrase)
			}
			if !skip(tags) {
				if _, found := get(commands, phrase); found {
					fatal(f.Name() + ": phrase already defined: " + phrase)
				}
				commands = append(commands, Command{phrase, tags, action})
			}
		}
		if sc.Err() != nil {
			fatal(sc.Err())
		}
	}
	return commands
}

func checkAudio(r io.ReadSeeker) uint32 {
	d := wav.NewDecoder(r)
	d.ReadInfo()
	if d.NumChans != 1 || d.WavAudioFormat != 1 {
		panic("audio must be the WAV format and mono")
	}
	return d.SampleRate
}

type EventType int
const (
	ResetEvent EventType = iota
	TranscribeEvent
	RapidOnEvent
	RapidOffEvent
)
type Event struct {
	Type EventType
	Content any
}

func handleFinalized(cmds []Command, phrases []string) []Event {
	var events []Event
	for _, p := range phrases {
		c, _ := get(cmds, p)
		action := c.Action
		var e []Event
		for _, t := range c.Tags {
			switch t {
			case "cancel":
				fmt.Println(action)
				return e
			case "transcribe":
				 e = append(e, Event{TranscribeEvent, action})
				 action = ""
			case "rapidon":
				e = append(e, Event{RapidOnEvent, nil})
			case "rapidoff":
				e = append(e, Event{RapidOffEvent, nil})
			}
		}
		events = append(events, e...)
		fmt.Println(action)
	}
	return events
}

func handleUnfinalized(cmds []Command, phrases []string, rapid bool) []Event {
	if len(phrases) == 0 {
		return nil
	}
	c, _ := get(cmds, phrases[len(phrases)-1])
	instant := false
	cancel := false
	for _, t := range c.Tags {
		if t == "instant" {
			instant = true
		} else if t == "cancel" {
			cancel = true
		}
	}
	if instant || rapid {
		if cancel {
			handleFinalized(cmds, []string{phrases[len(phrases)-1]})
			return []Event{Event{ResetEvent, nil}}
		}
		events := handleFinalized(cmds, phrases)
		return append(events, Event{ResetEvent, nil})
	}
	return nil
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
				for _, h := range []string{"kernel", "x11"} {
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
		sampleRate := float64(checkAudio(os.Stdin))
		phrases := make([]string, len(commands))
		for i := range commands {
			phrases[i] = commands[i].Phrase
		}
		cmdRec, err = vox.NewRecognizer(model, sampleRate, phrases)
		if err != nil {
			fatal(err)
		}
		cmdRec.SetWords(true)

		transRec, err = vox.NewRecognizer(model, sampleRate, nil)
		if err != nil {
			fatal(err)
		}
		transRec.SetMaxAlternatives(10)
	}

	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 4096)

	commanding := true
	rapid := false
	var transcriptAction string
	for {
		_, err :=  io.ReadFull(r, buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			panic(err.Error())
		}

		if commanding {
			var events []Event
			finalized, err := cmdRec.Accept(buf)
			if err != nil {
				panic(err.Error())
			}
			if finalized {
				result := cmdRec.Results()[0]
				phrases := strings.Fields(result.Text)
				events = handleFinalized(commands, phrases)
			} else {
				result := cmdRec.Results()[0]
				phrases := strings.Fields(result.Text)
				events = handleUnfinalized(commands, phrases, rapid)
			}

			for _, e := range events {
				switch e.Type {
				case ResetEvent:
					cmdRec.Purge()
				case TranscribeEvent:
					commanding = false
					transcriptAction = e.Content.(string)
					// Seems partials are output the audio chunk after they end,
					// so we should feed it to the other recognizer.
					_, err := transRec.Accept(buf)
					if err != nil {
						panic(err.Error())
					}
				case RapidOnEvent:
					rapid = true
				case RapidOffEvent:
					rapid = false
				}
			}
		} else {
			finalized, err := transRec.Accept(buf)
			if err != nil {
				panic(err.Error())
			}
			if finalized {
				for i, result := range transRec.Results() {
					fmt.Printf("transcript%d:%s\n", i+1, result.Text)
				}
				fmt.Println(transcriptAction)
				cmdRec.Purge()
				commanding = true
			}
		}
	}

	// We don't handle any final bit of audio.
}
