// +build windows

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
	config.textEdit.AppendText(text + "\r\n")
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
					if config.cmd != nil {
						log.Println("current pid is ", config.cmd.Process.Pid)
						config.logToTextarea("[kcptun] already running.")
						return
					}
					go func() {
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
				ReadOnly: true,
				VScroll: true,
			},
		},
	}).Run(); err != nil {
		log.Fatal(err)
	}
}
