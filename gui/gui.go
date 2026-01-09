package gui

import (
	"fmt"
	"net/http"

	"lcc-go/server"
	"lcc-go/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var httpServer *http.Server

func StartGUI(a fyne.App) {
	w := a.NewWindow("Buzzer Control Panel")

	ip := utils.GetLocalIP()
	status := widget.NewLabel("Status: STOPPED")

	startBtn := widget.NewButton("START SERVER", func() {
		if httpServer != nil {
			return
		}
		httpServer = server.StartServer()
		status.SetText(fmt.Sprintf("Status: RUNNING\nIP: %s:3000", ip))
	})

	stopBtn := widget.NewButton("STOP SERVER", func() {
		if httpServer == nil {
			return
		}
		httpServer.Close()
		httpServer = nil
		status.SetText("Status: STOPPED")
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("CERDAS CERMAT BUZZER"),
		status,
		startBtn,
		stopBtn,
	))

	w.Resize(fyne.NewSize(320, 200))
	w.ShowAndRun()
}
