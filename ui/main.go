// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"log"
	"os/exec"
	"sync"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var te *walk.TextEdit

	if _, err := (MainWindow{
		Title:  "Walk Clipboard Example",
		Size:   Size{300, 500},
		Layout: VBox{},
		Children: []Widget{
			PushButton{
				Text: "Copy",
				OnClicked: func() {
					if err := walk.Clipboard().SetText(te.Text()); err != nil {
						log.Print("Copy: ", err)
					}
				},
			},
			PushButton{
				Text: "Paste",
				OnClicked: func() {
					if text, err := walk.Clipboard().Text(); err != nil {
						log.Print("Paste: ", err)
					} else {
						te.SetText(text)
					}
				},
			},
			PushButton{
				Text: "Run",
				OnClicked: func() {
					go func() {
						runCmd(te)
					}()
				},
			},
			TextEdit{
				AssignTo: &te,
			},
		},
	}).Run(); err != nil {
		log.Fatal(err)
	}
}

func runCmd(te *walk.TextEdit) {
	var wg sync.WaitGroup
	bin := "C:\\Users\\alcheung\\programs\\kcptun\\kcptun.exe"
	args := []string{"-c", "C:\\Users\\alcheung\\programs\\kcptun\\bwg.json"}
	cmd := exec.Command(bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	outScanner := bufio.NewScanner(stdout)
	errScanner := bufio.NewScanner(stderr)

	wg.Add(2)
	go func() {
		defer wg.Done()
		for outScanner.Scan() {
			text := outScanner.Text()
			log.Println(text)
			if len(text) > 0 {
				te.AppendText(text + "\n")
			}
		}
	}()

	go func() {
		defer wg.Done()
		for errScanner.Scan() {
			text := errScanner.Text()
			log.Println(text)
			if len(text) > 0 {
				te.AppendText(text + "\n")
			}
		}
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
