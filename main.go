package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"mercury-client/modem"
)

type AppState struct {
	modemClient *modem.ModemClient

	myCallsignEntry     *widget.Entry
	targetCallsignEntry *widget.Entry
	targetIPEntry       *widget.Entry
	arqPortEntry        *widget.Entry
	broadcastPortEntry  *widget.Entry

	arqMessageEntry       *widget.Entry
	broadcastMessageEntry *widget.Entry
	filePathEntry         *widget.Entry
	selectFileButton      *widget.Button

	connectButton       *widget.Button
	disconnectButton    *widget.Button
	connectARQButton    *widget.Button
	disconnectARQButton *widget.Button
	abortARQButton      *widget.Button
	sendARQMsgButton    *widget.Button
	sendARQFileButton   *widget.Button
	sendBroadcastButton *widget.Button

	logOutput *widget.Entry

	chatOutput  *widget.Entry
	remoteCall  string
	chatRxBuffer string

	mainWin fyne.Window
}

func main() {
	a := app.New()
	w := a.NewWindow("Mercury Client")

	appState := &AppState{}
	appState.mainWin = w

	appState.setupUI(a)

	w.SetContent(appState.createContent())
	w.Resize(fyne.NewSize(800, 600))
	w.SetOnClosed(func() {
		if appState.modemClient != nil {
			appState.modemClient.Disconnect()
		}
	})
	w.ShowAndRun()
}

func (appState *AppState) setupUI(a fyne.App) {
	appState.myCallsignEntry = widget.NewEntry()
	appState.myCallsignEntry.SetPlaceHolder("My Callsign")
	appState.myCallsignEntry.SetText("NOCALL")

	appState.targetCallsignEntry = widget.NewEntry()
	appState.targetCallsignEntry.SetPlaceHolder("Target Callsign")
	appState.targetCallsignEntry.SetText("DEST")

	appState.targetIPEntry = widget.NewEntry()
	appState.targetIPEntry.SetPlaceHolder("IP Address")
	appState.targetIPEntry.SetText("127.0.0.1")

	appState.arqPortEntry = widget.NewEntry()
	appState.arqPortEntry.SetPlaceHolder("ARQ Port")
	appState.arqPortEntry.SetText("8300")

	appState.broadcastPortEntry = widget.NewEntry()
	appState.broadcastPortEntry.SetPlaceHolder("Broadcast Port")
	appState.broadcastPortEntry.SetText("8100")

	appState.arqMessageEntry = widget.NewMultiLineEntry()
	appState.arqMessageEntry.SetPlaceHolder("Enter ARQ message here...")

	appState.broadcastMessageEntry = widget.NewMultiLineEntry()
	appState.broadcastMessageEntry.SetPlaceHolder("Enter Broadcast message here...")

	appState.filePathEntry = widget.NewEntry()
	appState.filePathEntry.SetPlaceHolder("No file selected")
	appState.filePathEntry.Disable()

	appState.selectFileButton = widget.NewButton("Select File", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, appState.mainWin)
				return
			}
			if reader == nil {
				return
			}
			appState.filePathEntry.SetText(reader.URI().Path())
		}, appState.mainWin)
	})

	appState.connectButton = widget.NewButton("Connect TCP", appState.connect)
	appState.disconnectButton = widget.NewButton("Disconnect TCP", appState.disconnect)
	appState.disconnectButton.Disable()

	appState.connectARQButton = widget.NewButton("Connect", appState.connectARQ)
	appState.connectARQButton.Disable()
	appState.disconnectARQButton = widget.NewButton("Disconnect", appState.disconnectARQ)
	appState.disconnectARQButton.Disable()
	appState.abortARQButton = widget.NewButton("Abort", appState.abortARQ)
	appState.abortARQButton.Disable()

	appState.sendARQMsgButton = widget.NewButton("Send ARQ Message", appState.sendARQMessage)
	appState.sendARQMsgButton.Disable()

	appState.sendARQFileButton = widget.NewButton("Send ARQ File", appState.sendARQFile)
	appState.sendARQFileButton.Disable()

	appState.sendBroadcastButton = widget.NewButton("Send Broadcast (KISS)", appState.sendBroadcast)
	appState.sendBroadcastButton.Disable()

	appState.logOutput = widget.NewMultiLineEntry()
	appState.logOutput.SetPlaceHolder("Activity Log...")
	appState.logOutput.Wrapping = fyne.TextWrapBreak
	appState.logOutput.Disable()

	appState.chatOutput = widget.NewMultiLineEntry()
	appState.chatOutput.SetPlaceHolder("No ARQ chat messages yet...")
	appState.chatOutput.Wrapping = fyne.TextWrapBreak
	appState.chatOutput.Disable()
}

func (appState *AppState) createContent() fyne.CanvasObject {
	configForm := widget.NewForm(
		&widget.FormItem{Text: "My Callsign", Widget: appState.myCallsignEntry},
		&widget.FormItem{Text: "Target Callsign", Widget: appState.targetCallsignEntry},
		&widget.FormItem{Text: "IP Address", Widget: appState.targetIPEntry},
		&widget.FormItem{Text: "ARQ Port", Widget: appState.arqPortEntry},
		&widget.FormItem{Text: "Broadcast Port", Widget: appState.broadcastPortEntry},
	)

	connectionButtons := container.NewHBox(
		appState.connectButton,
		appState.disconnectButton,
	)

	arqButtons := container.NewHBox(
		appState.connectARQButton,
		appState.disconnectARQButton,
		appState.abortARQButton,
	)

	messageInput := container.NewVBox(
		widget.NewLabel("ARQ Message (sent over data port):"),
		appState.arqMessageEntry,
		appState.sendARQMsgButton,
	)

	fileInput := container.NewVBox(
		widget.NewLabel("ARQ File (sent over data port):"),
		container.NewBorder(
			nil, nil, nil, appState.selectFileButton, appState.filePathEntry,
		),
		appState.sendARQFileButton,
	)

	broadcastInput := container.NewVBox(
		widget.NewLabel("Broadcast Message (KISS):"),
		appState.broadcastMessageEntry,
		appState.sendBroadcastButton,
	)

	controls := container.NewVBox(
		widget.NewLabel("TCP Connection:"),
		connectionButtons,
		widget.NewLabel("ARQ Session:"),
		arqButtons,
		widget.NewSeparator(),
		messageInput,
		widget.NewSeparator(),
		fileInput,
		widget.NewSeparator(),
		broadcastInput,
	)

	leftPanel := container.NewVBox(
		configForm,
		layout.NewSpacer(),
		controls,
	)

	chatBox := container.NewBorder(
		widget.NewLabelWithStyle("ARQ Chat", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil, container.NewScroll(appState.chatOutput),
	)

	activityLogBox := container.NewBorder(
		widget.NewLabelWithStyle("Activity Log", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil, container.NewScroll(appState.logOutput),
	)

	rightPanel := container.NewBorder(
		nil, nil, nil, nil,
		container.NewVSplit(chatBox, activityLogBox),
	)

	return container.NewHSplit(leftPanel, rightPanel)
}

func (appState *AppState) connect() {
	targetIP := appState.targetIPEntry.Text
	arqPort := appState.arqPortEntry.Text
	broadcastPort := appState.broadcastPortEntry.Text

	arqControlAddr := fmt.Sprintf("%s:%s", targetIP, arqPort)

	arqPortNum, err := strconv.Atoi(arqPort)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid ARQ port: %v", err), appState.mainWin)
		return
	}
	arqDataAddr := fmt.Sprintf("%s:%d", targetIP, arqPortNum+1)
	broadcastAddr := fmt.Sprintf("%s:%s", targetIP, broadcastPort)

	appState.modemClient = modem.NewModemClient(arqControlAddr, arqDataAddr, broadcastAddr)

	err = appState.modemClient.Connect()
	if err != nil {
		dialog.ShowError(err, appState.mainWin)
		appState.logMessage(fmt.Sprintf("Connection error: %v", err))
		return
	}

	appState.logMessage("Connected to modem.")
	appState.setTCPConnected(true)

	appState.modemClient.SendCommand("MYCALL " + appState.myCallsignEntry.Text)
	appState.modemClient.SendCommand("LISTEN ON")
	appState.modemClient.SendCommand("PUBLIC OFF")
	appState.modemClient.SendCommand("COMPRESSION OFF")
	appState.modemClient.SendCommand("BW2750")

	go appState.updateLog()
	go appState.handleIncomingARQ()
	go appState.handleIncomingARQData()
	go appState.handleIncomingBroadcast()
	go appState.handleStatus()
}

func (appState *AppState) disconnect() {
	if appState.modemClient != nil {
		appState.modemClient.Disconnect()
		appState.modemClient = nil
	}
	appState.logMessage("Disconnected from modem.")
	appState.setTCPConnected(false)
	appState.setARQConnected(false)
}

func (appState *AppState) connectARQ() {
	if appState.modemClient == nil || !appState.modemClient.IsConnected() {
		dialog.ShowError(fmt.Errorf("not connected to modem"), appState.mainWin)
		return
	}

	src := appState.myCallsignEntry.Text
	dst := appState.targetCallsignEntry.Text
	appState.logMessage(fmt.Sprintf("Connecting ARQ: %s -> %s", src, dst))

	go func() {
		err := appState.modemClient.ConnectARQ(src, dst)
		if err != nil {
			appState.logMessage(fmt.Sprintf("ARQ connect error: %v", err))
			appState.setARQConnected(false)
			return
		}
		appState.logMessage(fmt.Sprintf("ARQ connected: %s -> %s", src, dst))
		appState.setARQConnected(true)
	}()
}

func (appState *AppState) disconnectARQ() {
	if appState.modemClient == nil {
		return
	}
	appState.logMessage("Disconnecting ARQ session...")
	appState.modemClient.DisconnectARQ()
	appState.setARQConnected(false)
}

func (appState *AppState) abortARQ() {
	if appState.modemClient == nil {
		return
	}
	appState.logMessage("Aborting ARQ session...")
	appState.modemClient.AbortARQ()
	appState.setARQConnected(false)
}

func (appState *AppState) sendARQMessage() {
	msg := appState.arqMessageEntry.Text
	if msg == "" {
		dialog.ShowInformation("Empty Message", "Please enter a message to send.", appState.mainWin)
		return
	}

	if appState.modemClient == nil || !appState.modemClient.IsConnected() {
		dialog.ShowError(fmt.Errorf("not connected to modem"), appState.mainWin)
		return
	}

	err := appState.modemClient.SendARQData([]byte(msg + "\n"))
	if err != nil {
		dialog.ShowError(err, appState.mainWin)
		appState.logMessage(fmt.Sprintf("Error sending ARQ data: %v", err))
		return
	}
	appState.logMessage(fmt.Sprintf("ARQ Data TX: %d bytes", len(msg)))
	appState.appendChat(fmt.Sprintf("%s: %s", appState.myCallsignEntry.Text, msg))
	appState.arqMessageEntry.SetText("")
}

func (appState *AppState) sendARQFile() {
	filePath := appState.filePathEntry.Text
	if filePath == "" || filePath == "No file selected" {
		dialog.ShowInformation("No File Selected", "Please select a file to send.", appState.mainWin)
		return
	}

	if appState.modemClient == nil || !appState.modemClient.IsConnected() {
		dialog.ShowError(fmt.Errorf("not connected to modem"), appState.mainWin)
		return
	}

	go func() {
		err := appState.modemClient.SendARQFile(filePath)
		if err != nil {
			appState.logMessage(fmt.Sprintf("Error sending ARQ file: %v", err))
			return
		}
		appState.logMessage("ARQ file transfer complete.")
	}()
}

func (appState *AppState) sendBroadcast() {
	msg := appState.broadcastMessageEntry.Text
	if msg == "" {
		dialog.ShowInformation("Empty Message", "Please enter a message to broadcast.", appState.mainWin)
		return
	}

	if appState.modemClient == nil || !appState.modemClient.IsConnected() {
		dialog.ShowError(fmt.Errorf("not connected to modem"), appState.mainWin)
		return
	}

	err := appState.modemClient.SendBroadcast([]byte(msg))
	if err != nil {
		dialog.ShowError(err, appState.mainWin)
		appState.logMessage(fmt.Sprintf("Error sending Broadcast message: %v", err))
	}
	appState.broadcastMessageEntry.SetText("")
}

func (appState *AppState) setTCPConnected(connected bool) {
	fyne.Do(func() {
		if connected {
			appState.connectButton.Disable()
			appState.disconnectButton.Enable()
			appState.connectARQButton.Enable()
			appState.sendBroadcastButton.Enable()
		} else {
			appState.connectButton.Enable()
			appState.disconnectButton.Disable()
			appState.connectARQButton.Disable()
			appState.disconnectARQButton.Disable()
			appState.abortARQButton.Disable()
			appState.sendARQMsgButton.Disable()
			appState.sendARQFileButton.Disable()
			appState.sendBroadcastButton.Disable()
		}
	})
}

func (appState *AppState) setARQConnected(connected bool) {
	fyne.Do(func() {
		if connected {
			appState.connectARQButton.Disable()
			appState.disconnectARQButton.Enable()
			appState.abortARQButton.Enable()
			appState.sendARQMsgButton.Enable()
			appState.sendARQFileButton.Enable()
		} else {
			appState.connectARQButton.Enable()
			appState.disconnectARQButton.Disable()
			appState.abortARQButton.Disable()
			appState.sendARQMsgButton.Disable()
			appState.sendARQFileButton.Disable()
		}
	})
}

func (appState *AppState) logMessage(msg string) {
	fyne.Do(func() {
		ts := time.Now().Format("15:04:05")
		line := fmt.Sprintf("[%s] %s", ts, msg)
		currentText := appState.logOutput.Text
		var newText string
		if currentText == "" {
			newText = line
		} else {
			newText = fmt.Sprintf("%s\n%s", line, currentText)
		}
		appState.logOutput.SetText(newText)
		appState.logOutput.Refresh()
	})
}

func (appState *AppState) updateLog() {
	if appState.modemClient == nil {
		return
	}
	for logMsg := range appState.modemClient.LogCh {
		appState.logMessage(logMsg)
	}
}

func (appState *AppState) handleIncomingARQ() {
	if appState.modemClient == nil {
		return
	}
	for arqMsg := range appState.modemClient.IncomingARQCh {
		appState.logMessage(fmt.Sprintf("ARQ Control: %s", arqMsg))
		if strings.HasPrefix(arqMsg, "CONNECTED") {
			appState.updateRemoteCall(arqMsg)
		}
	}
}

func (appState *AppState) updateRemoteCall(connectedLine string) {
	fields := strings.Fields(connectedLine)
	if len(fields) < 3 {
		return
	}
	myCall := strings.ToUpper(strings.TrimSpace(appState.myCallsignEntry.Text))
	callA := fields[1]
	callB := fields[2]
	if strings.ToUpper(callA) == myCall {
		appState.remoteCall = callB
	} else if strings.ToUpper(callB) == myCall {
		appState.remoteCall = callA
	} else {
		appState.remoteCall = callA
	}
	appState.logMessage(fmt.Sprintf("Remote ARQ callsign set to: %s", appState.remoteCall))
}

func (appState *AppState) handleIncomingARQData() {
	if appState.modemClient == nil {
		return
	}
	for data := range appState.modemClient.IncomingARQDataCh {
		appState.logMessage(fmt.Sprintf("ARQ Data RX: %d bytes: %q", len(data), string(data)))
		appState.chatRxBuffer += string(data)
		for {
			idx := strings.IndexByte(appState.chatRxBuffer, '\n')
			if idx < 0 {
				break
			}
			line := strings.TrimRight(appState.chatRxBuffer[:idx], "\r")
			appState.chatRxBuffer = appState.chatRxBuffer[idx+1:]
			if strings.TrimSpace(line) != "" {
				call := appState.remoteCall
				if call == "" {
					call = appState.targetCallsignEntry.Text
				}
				appState.appendChat(fmt.Sprintf("%s: %s", call, line))
			}
		}
		if len(appState.chatRxBuffer) > 65536 {
			appState.chatRxBuffer = appState.chatRxBuffer[len(appState.chatRxBuffer)-4096:]
		}
	}
}

func (appState *AppState) appendChat(line string) {
	fyne.Do(func() {
		currentText := appState.chatOutput.Text
		if currentText == "" {
			appState.chatOutput.SetText(line)
		} else {
			appState.chatOutput.SetText(fmt.Sprintf("%s\n%s", line, currentText))
		}
		appState.chatOutput.Refresh()
	})
}

func (appState *AppState) handleIncomingBroadcast() {
	if appState.modemClient == nil {
		return
	}
	for bcastData := range appState.modemClient.IncomingBcastCh {
		appState.logMessage(fmt.Sprintf("Broadcast RX (Decoded): %s", string(bcastData)))
	}
}

func (appState *AppState) handleStatus() {
	if appState.modemClient == nil {
		return
	}
	for status := range appState.modemClient.StatusCh {
		appState.logMessage(fmt.Sprintf("TNC Status: %s", status))

		switch status {
		case "CONNECTED":
			appState.setARQConnected(true)
		case "DISCONNECTED":
			appState.remoteCall = ""
			appState.chatRxBuffer = ""
			appState.setARQConnected(false)
		}
	}
}
