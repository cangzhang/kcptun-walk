// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
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
						if err := te.SetText(text); err != nil {
							log.Println(err)
						}
					}
				},
			},
			PushButton{
				Text: "Run",
				OnClicked: func() {
					go func() {
						log.Println("start...")
						start(te)
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
