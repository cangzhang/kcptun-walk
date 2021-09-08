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
	mainWin  *walk.MainWindow
	tray     *walk.NotifyIcon
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
	config.mainWin = mainW

	icon, err := walk.Resources.Icon("./assets/icon.ico")
	if err != nil {
		log.Fatal(err)
	}

	mainWConfig := MainWindow{
		AssignTo: &mainW,
		Title:    " Kcptun Walk",
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

	tray, err := walk.NewNotifyIcon(mainW)
	if err != nil {
		log.Fatal(err)
	}
	defer tray.Dispose()

	if err := tray.SetIcon(icon); err != nil {
		log.Fatal(err)
	}

	if err := tray.SetToolTip("App started"); err != nil {
		log.Fatal(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	tray.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}

		config.toggleVisible()
	})

	// toggle visible
	config.addTrayAction("T&oggle visible", func() {
		config.toggleVisible()
	})
	// exit
	config.addTrayAction("E&xit", func() {
		walk.App().Exit(0)
	})

	if err := tray.SetVisible(true); err != nil {
		log.Fatal(err)
	}

	config.tray = tray
	// Now that the icon is visible, we can bring up an info balloon.
	//if err := tray.ShowInfo("App started", "Click the icon to toggle visibility"); err != nil {
	//	log.Fatal(err)
	//}

	if _, err := mainWConfig.Run(); err != nil {
		log.Fatal(err)
	}
}

func (config *Config) toggleVisible() {
	if config.mainWin.Visible() {
		config.mainWin.Hide()
	} else {
		config.mainWin.Show()
	}
}

func (config *Config) addTrayAction(text string, cb func()) {
	action := walk.NewAction()
	if err := action.SetText(text); err != nil {
		log.Fatal(err)
	}
	action.Triggered().Attach(func() {
		cb()
	})
	if err := config.tray.ContextMenu().Actions().Add(action); err != nil {
		log.Fatal(err)
	}
}
