// +build windows

package main

import (
	"bytes"
	_ "embed"
	"image/png"
	"log"
	"os/exec"

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

var (
	tagName = "dev"
	sha     = "0000000"
)

//go:embed assets/icon.png
var appIcon []byte

func (config *Config) logToTextarea(text string) {
	log.Println(text)
	config.textEdit.AppendText(text + "\r\n")
}

func main() {
	var config Config
	mainW, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	img, err := png.Decode(bytes.NewReader(appIcon))
	if err != nil {
		log.Fatal(err)
	}
	icon, err := walk.NewIconFromImageForDPI(img, 96)
	if err != nil {
		log.Fatal(err)
	}

	mainWConfig := MainWindow{
		AssignTo: &mainW,
		Title:    " Kcptun Walk ",
		Size:     Size{300, 500},
		Layout:   VBox{},
		Icon:     icon,
		Children: []Widget{
			TextLabel{
				Text: "Version " + tagName + "(" + sha + ")",
			},
			PushButton{
				Text: "Run",
				OnClicked: func() {
					if config.cmd != nil {
						config.logToTextarea("current pid is " + string(rune(config.cmd.Process.Pid)))
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

		toggleVisible(mainW)
	})

	// toggle visible
	addTrayAction(tray, "T&oggle visible", func() {
		toggleVisible(mainW)
	})
	// exit
	addTrayAction(tray, "E&xit", func() {
		walk.App().Exit(0)
	})

	if err := tray.SetVisible(true); err != nil {
		log.Fatal(err)
	}

	config.tray = tray
	config.mainWin = mainW

	if _, err := mainWConfig.Run(); err != nil {
		log.Fatal(err)
	}
}

func toggleVisible(mainWin *walk.MainWindow) {
	if mainWin.Visible() {
		mainWin.Hide()
	} else {
		mainWin.Show()
	}
}

func addTrayAction(tray *walk.NotifyIcon, text string, cb func()) {
	if tray == nil {
		return
	}

	action := walk.NewAction()
	if err := action.SetText(text); err != nil {
		log.Fatal(err)
	}
	action.Triggered().Attach(func() {
		cb()
	})
	if err := tray.ContextMenu().Actions().Add(action); err != nil {
		log.Fatal(err)
	}
}
