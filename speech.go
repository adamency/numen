// sr MODEL_DIR PHRASE_FILE [PHRASE_FILE...]
package main

// TODO Carry over audio when switching recognizers.
// TODO Confidence threshold for @instant words.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/m7shapan/njson"
	"github.com/go-audio/wav"
	vosk "github.com/alphacep/vosk-api/go"
)

// Cut slices s around the first instance of sep,
// returning the text before and after sep.
// The found result reports whether sep appears in s.
// If sep does not appear in s, cut returns s, "", false.
func strings_Cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

func fatal(v ...interface{}) {
	fmt.Fprint(os.Stderr, "numen: ")
	fmt.Fprintln(os.Stderr, v...)
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
			phrase, action, found := strings_Cut(sc.Text(), ":")
			if len(action) > 0 {
				for []rune(action)[len([]rune(action))-1] == '\\' {
					if !sc.Scan() {
						fatal(f.Name() + ": unexpected end of file")
					}
					action = action[:len(action)-1] + "\n" + sc.Text()
				}
			}
			if !found {
				phrase, _, _ = strings_Cut(phrase, "#")
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

func getPhrases(json string) []string {
	var s struct {
		Words []string `njson:"result.#.word"`
		Partial string `njson:"partial"`
	}
	err := njson.Unmarshal([]byte(json), &s)
	if err != nil {
		panic(err)
	}
	if len(s.Partial) > 0 {
		return strings.Fields(s.Partial)
	}
	return s.Words
}

func getTranscripts(json string) []string {
	var s struct {
		Alternatives []string `njson:"alternatives.#.text"`
	}
	err := njson.Unmarshal([]byte(json), &s)
	if err != nil {
		panic(err)
	}
	return s.Alternatives
}

type Event int
const (
	ResetEvent Event = iota
	TranscribeEvent
	RapidOnEvent
	RapidOffEvent
)

func handleFinalized(cmds []Command, phrases []string) []Event {
	var events []Event
	for _, p := range phrases {
		c, _ := get(cmds, p)
		var e []Event
		for _, t := range c.Tags {
			switch t {
			case "cancel":
				fmt.Println(c.Action)
				return e
			case "transcribe":
				e = append(e, TranscribeEvent)
			case "rapidon":
				e = append(e, RapidOnEvent)
			case "rapidoff":
				e = append(e, RapidOffEvent)
			}
		}
		events = append(events, e...)
		fmt.Println(c.Action)
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
			return []Event{ResetEvent}
		}
		events := handleFinalized(cmds, phrases)
		return append(events, ResetEvent)
	}
	return nil
}

func reset(r *vosk.VoskRecognizer) {
	silence := make([]byte, 4096)
	r.AcceptWaveform(silence)
	r.Reset()
	r.AcceptWaveform(silence)
}

func main() {
	vosk.SetLogLevel(-1)
	model, err := vosk.NewModel(os.Args[1])
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

	var cmdRec, transRec *vosk.VoskRecognizer
	{
		sampleRate := float64(checkAudio(os.Stdin))

		var phrases strings.Builder
		phrases.WriteString("[")
		for _, c := range commands[:len(commands)-1] {
			phrases.WriteString(`"` + c.Phrase + `", `)
		}
		phrases.WriteString(`"` + commands[len(commands)-1].Phrase + `"]`)
		cmdRec, err = vosk.NewRecognizerGrm(model, sampleRate, phrases.String())
		if err != nil {
			fatal(err)
		}
		cmdRec.SetWords(1)

		transRec, err = vosk.NewRecognizer(model, sampleRate)
		if err != nil {
			fatal(err)
		}
		transRec.SetMaxAlternatives(10)
	}

	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 4096)

	commanding := true
	rapid := false
	for {
		_, err :=  io.ReadFull(r, buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			fatal(err)
		}

		if commanding {
			var events []Event
			if cmdRec.AcceptWaveform(buf) == 0 {
				events = handleUnfinalized(commands, getPhrases(cmdRec.PartialResult()), rapid)
			} else {
				events = handleFinalized(commands, getPhrases(cmdRec.FinalResult()))
				reset(cmdRec)
			}

			for _, e := range events {
				switch e {
				case ResetEvent:
					reset(cmdRec)
				case TranscribeEvent:
					commanding = false
				case RapidOnEvent:
					rapid = true
				case RapidOffEvent:
					rapid = false
				}
			}
		} else {
			if transRec.AcceptWaveform(buf) != 0 {
				for i, t := range getTranscripts(transRec.Result()) {
					fmt.Printf("transcript%d:%s\n", i+1, t)
				}
				reset(cmdRec)
				commanding = true
			}
		}
	}

	fmt.Println(string(cmdRec.FinalResult()))
}
