//filename: file_ops.go
//information: created on 3rd of November 2015 by Andreas Kittilsland
package main

import (
	"fmt"
	"gopkg.in/qml.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var FileLock = sync.Mutex{}

//Lock operations on file
func (ctrl *Control) Lock() {
	go func() {
		FileLock.Lock()
	}()
}

//Unlock operations on file
func (ctrl *Control) Unlock() {
	go func() {
		FileLock.Unlock()
	}()
}

//Save the file to disk
func (ctrl *Control) New_file() {
	go func() {
		ctrl.Log = "Creating clean document"
		qml.Changed(ctrl, &ctrl.Log)

		FileLock.Lock()

		ctrl.Plaintext = ""
		qml.Changed(ctrl, &ctrl.Plaintext)

		ctrl.Filename = "unsaved document"
		qml.Changed(ctrl, &ctrl.Filename)

		ctrl.Log = "New file created"
		qml.Changed(ctrl, &ctrl.Log)

		FileLock.Unlock()
	}()
}

//Save the file to disk
func (ctrl *Control) Save_file() {
	ctrl.Log = "Saving file..."
	qml.Changed(ctrl, &ctrl.Log)

	//Parse filename and directory
	fname := filepath.Base(ctrl.Savepath)
	dir := filepath.Dir(strings.Replace(ctrl.Savepath, "file:", "", 1))

	//Add file extension
	t := strings.Split(fname, ".")
	if len(t) == 1 || t[1] != "stxt" {
		fname = fname + ".stxt"
	}

	//Get and set working directory
	wd, err := os.Getwd()
	ErrorCheck(err, true)

	err = os.Chdir(dir)
	ErrorCheck(err, true)

	//Save file
	err = ioutil.WriteFile(fname, []byte(ctrl.Plaintext), 0644)
	if ErrorCheck(err) {
		ctrl.Log = "Failed to save file - " + string(err.Error())
		qml.Changed(ctrl, &ctrl.Log)
	} else {
		ctrl.Filename = fname
		qml.Changed(ctrl, &ctrl.Filename)

		ctrl.Log = "File was succesfully saved!"
		qml.Changed(ctrl, &ctrl.Log)
	}

	os.Chdir(wd)
	FileLock.Unlock()

	fmt.Println("changed")
}

//Open the file from disk
func (ctrl *Control) Open_file() {
	FileLock.Lock()

	//Open file
	fname := filepath.Base(ctrl.Openpath)
	dir := filepath.Dir(strings.Replace(ctrl.Openpath, "file:", "", 1))

	wd, err := os.Getwd()
	ErrorCheck(err, true)

	err = os.Chdir(dir)
	ErrorCheck(err, true)

	dat, err := ioutil.ReadFile(fname)
	if ErrorCheck(err) {
		ctrl.Log = "Failed to open file - " + string(err.Error())
		qml.Changed(ctrl, &ctrl.Log)
	} else {
		ctrl.Plaintext = string(dat)
		qml.Changed(ctrl, &ctrl.Plaintext)

		ctrl.Filename = fname
		qml.Changed(ctrl, &ctrl.Filename)

		ctrl.Log = "File was succesfully opened!"
		qml.Changed(ctrl, &ctrl.Log)
	}

	os.Chdir(wd)
	FileLock.Unlock()
}
