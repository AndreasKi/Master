//filename: gui.go
//information: created on 2nd of November 2015 by Andreas Kittilsland
package main

import (
	"gopkg.in/qml.v1"
	"os"
)

var GuiInterface *Control

//Close the application if the user exits the GUI
func (ctrl *Control) Exit() {
	go func() {
		os.Exit(0)
	}()
}

//Start the GUI
func run() error {
	engine := qml.NewEngine()

	controls, err := engine.LoadFile("graphics.qml")
	if err != nil {
		return err
	}

	ctrl := Control{Log: "No errors", Filename: "unsaved document"}

	context := engine.Context()
	context.SetVar("ctrl", &ctrl)
	GuiInterface = &ctrl

	window := controls.CreateWindow(nil)

	window.Show()
	ctrl.Root = window.Root()
	accepting.Unlock()
	window.Wait()

	return nil
}

type Control struct {
	Root              qml.Object
	Filename          string
	Log               string
	Plaintext         string
	Plainselectedtext string
	Plainbeforetext   string
	Plainaftertext    string
	Openpath          string
	Savepath          string
	Startselection    int
	Endselection      int
	Cursorposition    int
}
