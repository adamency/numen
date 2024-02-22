package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

type Audio struct {
	Device string
	cmd    *exec.Cmd
	stdout io.ReadCloser

	Filename string
	file     *os.File
}

func (a *Audio) SetDevice(device string) {
	if device == "" {
		out, _ := exec.Command("arecord", "-L").Output()
		if bytes.Contains(out, []byte("sysdefault:CARD=Microphone\n")) {
			device = "sysdefault:CARD=Microphone"
		} else {
			device = "default"
		}
	}
	a.Device = device
}

func (a *Audio) Start() error {
	if a.Filename == "" {
		a.cmd = exec.Command("arecord", "-q", "-fS16_LE", "-c1", "-r16000", "-D", a.Device)
		a.cmd.Stderr = os.Stderr
		var err error
		a.stdout, err = a.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		return a.cmd.Start()
	}

	var err error
	a.file, err = os.Open(a.Filename)
	return err
}

func (a *Audio) Reader() io.Reader {
	if a.stdout != nil {
		return a.stdout
	}
	return a.file
}

func (a *Audio) Close() error {
	if a.cmd != nil {
		return a.cmd.Wait()
	}
	return a.file.Close()
}
