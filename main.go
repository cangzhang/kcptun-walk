// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os/exec"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type Config struct {
	textEdit *walk.TextEdit
	binPath  string
	binDir   string
	jsonPath string
	pwd      string
	cmd      *exec.Cmd
}

func (config *Config) logToTextarea(text string) {
	config.textEdit.AppendText("\n" + text + "\n")
}

func main() {
	var config Config

	if _, err := (MainWindow{
		Title:  "Kcptun Walk",
		Size:   Size{300, 500},
		Layout: VBox{},
		Children: []Widget{
			PushButton{
				Text: "Run",
				OnClicked: func() {
					go func() {
						config.textEdit.SetReadOnly(true)
						startCmd(&config)
					}()
				},
			},
			PushButton{
				Text: "Stop",
				OnClicked: func() {
					go func() {
						killCmd(&config)
					}()
				},
			},
			TextEdit{
				AssignTo: &config.textEdit,
			},
		},
	}).Run(); err != nil {
		log.Fatal(err)
	}
}
