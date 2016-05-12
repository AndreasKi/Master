//filename: func_main.go
//information: created on 8th of September 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	ToScreen("Initializing variables...")
	read_configuration() //func_initialize.go

	ToScreen("Setting up daemon...")
	if len(os.Args) >= 2 {
		if os.Args[1] == "fixed" {
			auto_coord = false
		} else if os.Args[1] == "auto" {
			auto_coord = true
		} else {
			ErrorCheck(errors.New("Invalid argument: \""+os.Args[1]+"\". Should be either \"fixed\" or \"auto\""), true)
		}
	}
	if len(os.Args) == 3 {
		set_port_and_find_coordinator(os.Args[2])
	} else if len(os.Args) == 4 {
		set_port_and_coordinator(os.Args[2])
	} else if len(os.Args) < 3 {
		find_port_and_coordinator()
	} else {
		ErrorCheck(errors.New("Number of arguments was too high: "+strconv.Itoa(len(os.Args)-1)+". Need port number, and coordinator. F. Ex."+
			"\n<./daemon <fixed/auto> 8590 c> if coordinator, or <./daemon <fixed/auto> 8591> if not coordinator"+
			"\n<./daemon> assumes use of only localhost/127.0.0.1, and looks for port on its own."), true)
	}
	//Get PID
	pid_string = strconv.Itoa(os.Getpid())

	//Initialize maps and slices
	version_vector = make(map[int]int)
	locks = make(map[int]*ExpandedLock) //lock.go
	objects = make(map[int]object)
	coord_version_vector = make(map[int]int)

	//fmt.Println(ip_self + ":" + port + ":" + pid_string)

	std_dir = std_dir + "daemon_" + port + "/"
	daemon_dir, _ = os.Getwd()
	//CreateDumpFile()     //file-handling.go
	CheckFileDirExists() //file-handling.go

	ToScreen("Starting servers...")
	go listen()     //communication.go
	Connect()       //init.go
	go CheckFiles() //file-handling.go
	run_gui()       //gui.go
}

//A local update was made, push it to queue
func daemon_InstigateUpdate(id string, version string, attributes map[string]string, timer string, obj_type string) {
	msg := "inst update\n" + id + "\nnot\n" + version + "\n" + obj_type + "\n" + timer + "\n"
	for index, item := range attributes {
		msg = msg + index + " " + item + "\n"
	}
	msg = msg + "done\n"
	WaitingUpdates.Push(msg)

	return
}

//A local delete op was made, push it to queue
func daemon_InstigateDelete(id string, timer string) {
	msg := "inst delete\n" + id + "\nnot\n" + timer + "\n"
	WaitingUpdates.Push(msg)

	return
}

//refresh command was made, push updates to the coordinator
func daemon_PushUpdates() {
	if coordinator_changing.IsLocked() {
		fmt.Println("Attempted to push updates, but coordinator is being moved.")
		coordinator_changing.Lock()
		coordinator_changing.Unlock()
	}
	if !Synced() {
		fmt.Println("Attempted to push updates, but daemon is not Synced with coordinator...")
		for !Synced() { //Ensure that initial synchronization is done (init.go)
			time.Sleep(500 * time.Millisecond)
		}
	}

	for true {
		if WaitingUpdates.Peek() == "nil" {
			break
		} else {
			conn, err := Dial(coord_ip, coord_port)
			ErrorCheck(err, true)
			WriteLines(conn, WaitingUpdates.Pop())
		}
	}
	return
}

//Set coordinator
func BecomeCoordinator(first_coordinator bool) {
	if !first_coordinator {
		coordinator_changing.Lock()
	}
	SendToAllDaemons("coordinator change\n" + ip_to + "\n" + port + "\n")
	if coord_port != "0" {
		DialAndSend(coord_ip, coord_port, "coordinator change\n"+ip_to+"\n"+port+"\n")
	}

	if first_coordinator {
		is_coord = true
		ToScreen("> I am the first coordinator")
	}

	if coord_port != "0" {
		AddDaemon(coord_ip, coord_port)
	}

	coord_ip = ip_to
	coord_port = port

	//SendToAllDaemons("coordinator changed\n")
	return
}

//Start automatic coordinator selection
func AutomaticCoordinator() {
	SendToAllDaemons("evaluate coordinator\n")
	DialAndSend(coord_ip, coord_port, "evaluate coordinator\n")
}

//Send request for updating or adding an object to coordinator
func SendInstigationRequest(id int, version int, attributes map[string]string, obj_type string) error {
	//Lock the file on the coordinator if it is an update not an add
	if is_coord && id != -1 {
		if _, ok := locks[id]; ok {
			locks[id].Lock()
		} else {
			locks[id] = &ExpandedLock{}
			locks[id].Lock()
		}
	}

	metrics_times.Push("change")

	daemon_InstigateUpdate(strconv.Itoa(id), strconv.Itoa(version), attributes, time.Now().Format(time.RFC3339), obj_type) //daemon.go

	return nil
}

//Send request for deleting an object to coordinator
func SendDeletionRequest(id int) error {
	metrics_times.Push("change")

	daemon_InstigateDelete(strconv.Itoa(id), time.Now().Format(time.RFC3339)) //daemon.go

	return nil
}

//Finds and adds all applications on this device
func FindApplications() {
	local_apps_mu.Lock()

	_, previous_list, new_apps := RetrieveLocalApplications()

	local_apps_mu.Unlock()

	conn, err := Dial(coord_ip, coord_port)
	ErrorCheck(err, true)

	WriteLines(conn, "app list\n")

	local_apps_mu.ReadOnly()
	WriteSlice(conn, new_apps)

	conn_reader := bufio.NewReader(conn)
	if ReadLine(conn_reader) != "ack" {
		ErrorCheck(errors.New("Did not receive acknowledgement from coordinator"), true)
	}
	conn.Close()

	var removed_apps []string
	for _, app := range previous_list {
		if _, ok := AppIn_String(local_apps, app); !ok {
			removed_apps = append(removed_apps, app)
		}
	}

	local_apps_mu.Editable()
	if len(removed_apps) > 0 {
		conn, err = Dial(coord_ip, coord_port)
		ErrorCheck(err, true)

		WriteLines(conn, "app remove list\n")
		WriteSlice(conn, removed_apps)

		conn_reader = bufio.NewReader(conn)
		if ReadLine(conn_reader) != "ack" {
			ErrorCheck(errors.New("Did not receive acknowledgement from coordinator"), true)
		}
		conn.Close()
	}
}
