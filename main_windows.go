//go:build windows

package main

import (
	_ "embed"
	"github.com/getlantern/systray"
	"opensavecloudserver/constant"
	"opensavecloudserver/server"
	"os"
)

//go:generate go-winres make

//go:embed tray.ico
var icon []byte

func main() {
	go func() {
		InitCommon()
		server.Serve()
	}()
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("Open Save Cloud Server")
	systray.SetTooltip("Open Save Cloud Server")
	systray.AddMenuItem("Open Save Cloud", "").Disable()
	systray.AddMenuItem(constant.Version, "").Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the server")
	select {
	case <-mQuit.ClickedCh:
		quit()
	}
}

func quit() {
	systray.Quit()
	os.Exit(0)
}

func onExit() {
	systray.Quit()
}
