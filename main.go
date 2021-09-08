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
	mainW, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	icon, err := walk.Resources.Icon("./assets/icon.ico")
	if err != nil {
		log.Fatal(err)
	}

	mainWConfig := MainWindow{
		AssignTo: &mainW,
		Title:    "Kcptun Walk",
		Size:     Size{300, 500},
		Layout:   VBox{},
		Icon:     icon,
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
				VScroll:  true,
			},
		},
	}

	ni, err := walk.NewNotifyIcon(mainW)
	if err != nil {
		log.Fatal(err)
	}
	defer ni.Dispose()

	if err := ni.SetIcon(icon); err != nil {
		log.Fatal(err)
	}

	if err := ni.SetToolTip("Click for info or use the context menu to exit."); err != nil {
		log.Fatal(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}

		if err := ni.ShowCustom(
			"Walk NotifyIcon Example",
			"There are multiple ShowX methods sporting different icons.",
			icon); err != nil {

			log.Fatal(err)
		}
	})

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}
	if err := ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}
	// Now that the icon is visible, we can bring up an info balloon.
	if err := ni.ShowInfo("Walk NotifyIcon Example", "Click the icon to show again."); err != nil {
		log.Fatal(err)
	}

	if _, err := mainWConfig.Run(); err != nil {
		log.Fatal(err)
	}
}
