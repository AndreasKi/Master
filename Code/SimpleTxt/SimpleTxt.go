//filename: SimpleTxt.go
//information: created on 2nd of November 2015 by Andreas Kittilsland
package main

import (
	"fmt"
	"gopkg.in/qml.v1"
	"os"
	"strconv"
	"sync"
)

var accepting = sync.Mutex{}

func main() {
	defer CloseApp()
	accepting.Lock()

	//Check if we were executed by a daemon wanting remote access
	if len(os.Args) == 2 {
		//Set up comms for remote interface as it was requested
		//One instance of the application is run for every object opened.
		id, err := strconv.Atoi(os.Args[1])
		ErrorCheck(err, true)

		port_int, err := strconv.Atoi(port)
		ErrorCheck(err, true)

		port_int = port_int + id

		port = strconv.Itoa(port_int)
	}

	//Wait for remote work
	go listen() //communication.go

	//Run the editor
	if err := qml.Run(run); err != nil { //gui.go
		ErrorCheck(err, true)
	}
}

func CloseApp() {
	fmt.Println("exit")
}
