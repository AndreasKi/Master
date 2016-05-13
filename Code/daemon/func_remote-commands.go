//filename: func_remote-commands.go
//information: created on 5th of October 2015 by Andreas Kittilsland
//Contains specific communications functions
package main

import (
	"bufio"
	"errors"
	"os"
	"net"
	"strconv"
	"time"
)

//Serve connections
func rc_ConnectionHandler(conn net.Conn) {
	defer conn.Close()
	conn_reader := bufio.NewReader(conn)
	line := ReadLine(conn_reader)

	if line == "test conn" {
		//Test connection, return accept
		WriteLines(conn, "accepted")
		WriteLines(conn, pid_string)
	} else if line == "request state" {
		rc_RequestState(conn, conn_reader)
	} else if line == "daemon req est conn" {
		rc_DaemonReqEstConn(conn, conn_reader)
	} else if line == "daemon est conn" {
		rc_DaemonEstablishConnection(conn, conn_reader)
	} else if line == "inst update" {
		rc_InstUpdate(conn, conn_reader)
	} else if line == "inst delete" {
		rc_InstDelete(conn, conn_reader)
	} else if line == "update" {
		rc_Update(conn, conn_reader)
	} else if line == "delete" {
		rc_Delete(conn, conn_reader)
	} else if line == "refresh" {
		calleeType := ReadLine(conn_reader)
		rc_RequestRefresh(calleeType)
		WriteLines(conn, "ack")
	} else if line == "check v vector" {
		rc_CheckVVector()
		WriteLines(conn, "ack")
	} else if line == "refresh objs" {
		daemon_PushUpdates()
		WriteLines(conn, "ack")
	} else if line == "req objs" {
		rc_ReqObjs(conn, conn_reader)
	} else if line == "change name" {
		rc_ChangeName(conn, conn_reader)
	} else if line == "run chng dtctr" {
		rc_RunChangeTracker(conn, conn_reader)
	} else if line == "client open file" {
		rc_ClientOpenFile(conn, conn_reader)
	} else if line == "req file access" {
		rc_RequestFileAccess(conn, conn_reader)
	} else if line == "joining" {
		rc_Joining(conn, conn_reader)
	} else if line == "refresh app list" {
		rc_RefreshAppsList(conn, conn_reader)
	} else if line == "app list" {
		rc_AppList(conn, conn_reader)
	} else if line == "app remove list" {
		rc_AppRemoveList(conn, conn_reader)
	} else if line == "coordinator change" {
		rc_CoordinatorChange(conn, conn_reader)
	} else if line == "coordinator changed" {
		rc_CoordinatorChanged(conn, conn_reader)
	} else if line == "coord tasks done" {
		WriteLines(conn, "ack")
		is_coord = true
		coordinator_changing.Unlock()
		ToScreen("> I am the coordinator")
		ToScreen("> Status: has " + strconv.Itoa(len(daemons)+1) + " daemons in network.")
	} else if line == "new file" {
		ToChangeTracker("add\n" + "daemon_" + port + "/" + ReadLine(conn_reader) + "\n")
		WriteLines(conn, "ack")
	} else if line == "evaluate coordinator" {
		WriteLines(conn, "ack")
		go ChooseCoordinator()
	} else if line == "set vars" {
		rc_SetVars(conn, conn_reader)
	} else if line == "vars changed" {
		rc_VarsChanged(conn, conn_reader)
	} else {
		ErrorCheck(errors.New("Failed to understand message: " + line))
	}
	return
}

func rc_DaemonReqEstConn(conn net.Conn, conn_reader *bufio.Reader) {
	if coordinator_changing.IsLocked() {
		WriteLines(conn, "N/A")
	} else {
		if is_coord {
			WriteLines(conn, "ack")
			coord_DaemonEstablishConnection(conn, conn_reader) //coordinator.go
		} else {
			WriteLines(conn, "!coord", coord_ip, coord_port)
		}
	}
	return
}

func rc_DaemonEstablishConnection(conn net.Conn, conn_reader *bufio.Reader) {
	//Another daemon wishes to join,
	//get its details
	ip_d := ReadLine(conn_reader)
	port_d := ReadLine(conn_reader)

	//Add the daemon to our local list
	AddDaemon(ip_d, port_d) //mutators.go
	WriteLines(conn, "ack")
	return
}

func rc_RequestState(conn net.Conn, conn_reader *bufio.Reader) {
	WriteLines(conn, "accepted")
	WriteIntLine(conn, nr_local_objs)

	//Make sure we discard expired values before sending
	ResetMetrics()

	usage_m.Lock()
	WriteIntLine(conn, nr_of_runs)
	WriteIntLine(conn, nr_of_changes)
	usage_m.Unlock()

	isCharging, battPercent := CheckBattery()
	isCharging_s := "false"
	if isCharging == true {
		isCharging_s = "true"
	}

	WriteLines(conn, isCharging_s)
	WriteIntLine(conn, battPercent)

	return
}

func rc_RequestRefresh(calleeType string) {
	if is_coord {
		refreshing.Lock()
		coordinator_serving_atm++

		//Ensure that all daemons are synchronized
		SendToAllDaemons("check v vector\n")

		//and make them push updates
		SendToAllDaemons("refresh objs\n")
		daemon_PushUpdates()

		//Update applications list, but dont wait for it if user requested objects
		UpdateApps := func() {
			sendApplicationsTo_mu.Lock()
			for len(sendApplicationsTo) > 0 {
				message := "app list\n"
				for _, app := range applications {
					message = message + app.app_name + "\n"
				}
				message = message + "done\n"
				DialAndSend(sendApplicationsTo[0].ip, sendApplicationsTo[0].port, message)

				sendApplicationsTo[0] = sendApplicationsTo[len(sendApplicationsTo)-1]
				sendApplicationsTo = sendApplicationsTo[:len(sendApplicationsTo)-1]
			}
			sendApplicationsTo_mu.Unlock()
			coord_FindApplications()
			SendToAllDaemons("refresh app list\n")
		}
		if calleeType == "objs" {
			go UpdateApps()
		} else {
			UpdateApps()
		}

		coordinator_serving_atm--
		current_updates.Wait() //Wait for the updates to get pushed before returning
		refreshing.Unlock()
	} else {
		DialAndSend(coord_ip, coord_port, "refresh\n"+calleeType+"\n")
	}

	return
}

func rc_ReqObjs(conn net.Conn, conn_reader *bufio.Reader) {
	ip_d := ReadLine(conn_reader)
	port_d := ReadLine(conn_reader)
	v := ReadVersionVector(conn_reader)
	SendObjsInVector(v, ip_d, port_d)
	WriteLines(conn, "ack")

	return
}

func rc_CheckVVector() {
	coord_ver_vector_mu.ReadOnly()
	ver_vector_mu.ReadOnly()
	diff_vec := make(map[int]int)
	for id, cversion := range coord_version_vector {
		if lversion, ok := version_vector[id]; ok && lversion == cversion {
			continue
		} else {
			diff_vec[id] = 0
		}
	}
	RequestObjsInVector(diff_vec)
	coord_ver_vector_mu.Editable()
	ver_vector_mu.Editable()
	return
}

func rc_RefreshAppsList(conn net.Conn, conn_reader *bufio.Reader) {
	FindApplications()
	WriteLines(conn, "ack")
	return
}

func rc_InstUpdate(conn net.Conn, conn_reader *bufio.Reader) {
	//Read the content of the message
	id := ReadIntLine(conn_reader)

	//Lock the file on the coordinator if it is an update not an add
	if is_coord && id != -1 {
		if _, ok := locks[id]; ok {
			locks[id].Lock()
		} else {
			locks[id] = &ExpandedLock{}
			locks[id].Lock()
		}
	}

	toQueueOrNot := ReadLine(conn_reader)

	//read the remaining lines
	version := ReadLine(conn_reader)
	obj_type := ReadLine(conn_reader)
	timer := ReadLine(conn_reader)
	attributes := ReadAttributesMap(conn_reader)

	var err error
	err = nil
	if toQueueOrNot == "not" {
		if is_coord {
			err = coord_InstigateUpdate(strconv.Itoa(id), version, attributes, timer, obj_type) //coordinator.go
		} else {
			ErrorCheck(errors.New("Profane daemon received update request"), true)
		}
	} else {
		daemon_InstigateUpdate(strconv.Itoa(id), version, attributes, timer, obj_type) //daemon.go
	}

	if err != nil {
		ErrorCheck(err)
		WriteLines(conn, "sync fail")
	} else {
		WriteLines(conn, "success")
	}
	return
}

func rc_InstDelete(conn net.Conn, conn_reader *bufio.Reader) {
	//Read the content of the message
	id := ReadIntLine(conn_reader)
	toQueueOrNot := ReadLine(conn_reader)
	timer := ReadLine(conn_reader)

	var err error
	err = nil
	if toQueueOrNot == "not" {
		if is_coord {
			err = coord_InstigateDelete(strconv.Itoa(id), timer) //coordinator.go
		} else {
			ErrorCheck(errors.New("Profane daemon received deletion request"), true)
		}
	} else {
		daemon_InstigateDelete(strconv.Itoa(id), timer) //daemon.go
	}

	if err != nil {
		ErrorCheck(err)
		WriteLines(conn, "sync fail")
	} else {
		WriteLines(conn, "success")
	}
	return
}

func rc_Update(conn net.Conn, conn_reader *bufio.Reader) {
	//Read the content of the message
	id := ReadIntLine(conn_reader)
	version := ReadIntLine(conn_reader)
	obj_type := ReadLine(conn_reader)
	timer := ReadLine(conn_reader)
	attributes := ReadAttributesMap(conn_reader)

	if _, ok := locks[id]; ok {
		locks[id].Lock()
	} else {
		locks[id] = &ExpandedLock{}
		locks[id].Lock()
	}

	//Update our coordinator vector right away
	coord_ver_vector_mu.Lock()
	coord_version_vector[id] = version
	coord_ver_vector_mu.Unlock()
	_, err := AddObject(id, version, attributes, timer, obj_type) //mutators.go
	locks[id].Unlock()
	if err != nil {
		WriteLines(conn, "dnf")
		ErrorCheck(err, true)
	}
	WriteLines(conn, "ack")
	return
}

func rc_Delete(conn net.Conn, conn_reader *bufio.Reader) {
	//Read the content of the message
	id := ReadIntLine(conn_reader)
	timer := ReadLine(conn_reader)

	if _, ok := locks[id]; ok {
		locks[id].Lock()
	} else {
		locks[id] = &ExpandedLock{}
		locks[id].Lock()
	}

	//Update our coordinator vector right away
	coord_ver_vector_mu.Lock()
	delete(coord_version_vector, id)
	coord_ver_vector_mu.Unlock()

	err := RemoveObject(id, timer) //mutators.go
	if err != nil {
		locks[id].Unlock()
		WriteLines(conn, "dnf")
		ErrorCheck(err, true)
	}
	WriteLines(conn, "ack")
	return
}

func rc_RunChangeTracker(conn net.Conn, conn_reader *bufio.Reader) {
	isAsync := ReadLine(conn_reader)
	if isAsync == "async" {
		RunChangeTrackerScan(true)
	} else {
		RunChangeTrackerScan(false)
	}
	WriteLines(conn, "ack")
	return
}

func rc_ClientOpenFile(conn net.Conn, conn_reader *bufio.Reader) {
	for coordinator_changing.IsLocked() {
		time.Sleep(time.Millisecond * coordinator_transfer_time)
	}
	id := ReadIntLine(conn_reader)

	//Check that the object exists
	if _, found := objects[id]; !found {
		//Object was likely removed very recently
		WriteLines(conn, "sync conflict")
		return
	}
	locks[id].ReadOnly()
	if objects[id].file_path == "N/A" {
		//Get file holder to expose file in browser
		file_conn, err := Dial(coord_ip, coord_port)
		ErrorCheck(err, true)

		WriteLines(file_conn, "req file access")
		WriteIntLine(file_conn, id)
		file_reader := bufio.NewReader(file_conn)
		outcome := ReadLine(file_reader)
		if outcome == "ok" {
			url := ReadLine(file_reader)
			WriteLines(conn, "found", url)
		} else {
			WriteLines(conn, "nf")
			ErrorCheck(errors.New("Could not find file with id " + strconv.Itoa(id) + ", or file was already open."))
		}

		file_conn.Close()
	} else {
		WriteLines(conn, "ok")
		metrics_times.Push("run")

		//create url to file
		temp, err := strconv.Atoi(http_gui_port)
		ErrorCheck(err, true)
		file_port := strconv.Itoa(temp + target_size)
		ExposeFileInterface(id, file_port) //remote_access.go

		url := ip_to + ":" + file_port + "/" + strconv.Itoa(id)
		WriteLines(conn, url)
	}
	locks[id].Editable()

	return
}

func rc_RequestFileAccess(conn net.Conn, conn_reader *bufio.Reader) {
	id := ReadIntLine(conn_reader)

	//Check that the object exists
	if _, found := objects[id]; !found {
		//Object was likely removed during last scan
		WriteLines(conn, "sync conflict")
		return
	}
	locks[id].ReadOnly()
	if objects[id].file_path == "N/A" {
		if is_coord {
			coord_request_file_access(conn, conn_reader, id) //coordinator.go
		} else {
			WriteLines(conn, "nf")
		}
	} else {
		if !InterfaceIsRunning(id) {
			WriteLines(conn, "ok")

			metrics_times.Push("run")

			//create url to file
			temp, err := strconv.Atoi(http_gui_port)
			ErrorCheck(err, true)
			file_port := strconv.Itoa(temp + target_size)
			ExposeFileInterface(id, file_port) //remote_access.go

			url := ip_to + ":" + file_port + "/" + strconv.Itoa(id)
			WriteLines(conn, url)
		} else {
			WriteLines(conn, "sync conflict")
		}
	}
	locks[id].Editable()
	return
}

func rc_Joining(conn net.Conn, conn_reader *bufio.Reader) {
	if is_coord { //Inform the new daemon that this instance is the coordinator
		WriteLines(conn, "coordinator")
	} else {
		WriteLines(conn, "not coordinator")
	}
	return
}

//A list of applications incoming. If is coord, only contains additions. If profane, full list.
func rc_AppList(conn net.Conn, conn_reader *bufio.Reader) {
	rec_apps := ReadSlice(conn_reader)
	applications_mu.Lock()
	changed := false
	for _, app := range rec_apps {
		if i, ok := AppIn_Struct(applications, app); !ok {
			applications = append(applications, application{app, 1})
			changed = true
		} else if is_coord {
			applications[i].count++
		}
	}
	if is_coord && changed {
		go coord_update_app_list(applications)
	} else if !is_coord {
		//See if some apps were removed since last update
		var removed_apps []string
		for _, app := range applications {
			if _, ok := AppIn_String(rec_apps, app.app_name); !ok {
				removed_apps = append(removed_apps, app.app_name)
			}
		}
		for _, app := range removed_apps {
			if i, ok := AppIn_Struct(applications, app); ok {
				applications[i] = applications[len(applications)-1]
				applications = applications[:len(applications)-1]
			}
		}

		//See if there are additions
		for _, app := range rec_apps {
			if _, ok := AppIn_Struct(applications, app); !ok {
				applications = append(applications, application{app, 1})
			}
		}
	}
	applications_mu.Unlock()
	WriteLines(conn, "ack")
	return
}

//A list of applications that have been removed are incomming.
func rc_AppRemoveList(conn net.Conn, conn_reader *bufio.Reader) {
	rec_apps := ReadSlice(conn_reader)
	applications_mu.Lock()
	changed := false
	for _, app := range rec_apps {
		if i, ok := AppIn_Struct(applications, app); ok {
			applications[i].count--
			if applications[i].count < 1 {
				changed = true
				applications[i] = applications[len(applications)-1]
				applications = applications[:len(applications)-1]
			}
		}
	}
	applications_mu.Unlock()
	if is_coord && changed {
		coord_update_app_list(applications)
	}
	WriteLines(conn, "ack")
	return
}

//A new coordinator is being set up
func rc_CoordinatorChange(conn net.Conn, conn_reader *bufio.Reader) {
	coordinator_changing.Lock()

	new_coord_ip := ReadLine(conn_reader)
	new_coord_port := ReadLine(conn_reader)

	if is_coord {
		go coord_finish_tasks(new_coord_ip, new_coord_port)
	} else {
		//Add old coordinator to daemon list
		AddDaemon(coord_ip, coord_port)

		coord_ip = new_coord_ip
		coord_port = new_coord_port

		//Remove new coordinator from daemon list
		daemons_mu.Lock()
		for i, d := range daemons {
			if d.ip == coord_ip && d.port == coord_port {
				RemoveDaemon(i)
				break
			}
		}
		daemons_mu.Unlock()
	}

	WriteLines(conn, "ack")

	return
}

//The new coordinator was done being set up, and is now ready
func rc_CoordinatorChanged(conn net.Conn, conn_Reader *bufio.Reader) {
	coordinator_changing.Unlock()
	WriteLines(conn, "ack")
	ToScreen("> I am a profane daemon")
	return
}

//Change name of file
func rc_ChangeName(conn net.Conn, conn_reader *bufio.Reader) {
	pfname := ReadLine(conn_reader)
	fname := ReadLine(conn_reader)
	id := ReadIntLine(conn_reader)

	obj_mu.ReadOnly()
	if objects[id].file_path != "N/A" {
		wd := SetWD(daemon_dir)
		os.Rename("../files/daemon_"+port+"/"+pfname, "../files/daemon_"+port+"/"+fname)
		ResetWD(wd)
		ToChangeTracker("name_change\n" + "daemon_" + port + "/" + pfname + ":daemon_" + port + "/" + fname + "\n")
	} else if is_coord {
		SendToAllDaemons("change name\n"+pfname+"\n"+fname+"\n"+strconv.Itoa(id)+"\n")
	}
	obj_mu.Editable()

	WriteLines(conn, "ack")
	return
}

//The new coordinator was done being set up, and is now ready
func rc_SetVars(conn net.Conn, conn_reader *bufio.Reader) {
	interval_minutes := ReadLine(conn_reader)
	interval_seconds := ReadLine(conn_reader)

	weight_objects := ReadLine(conn_reader)
	weight_runs := ReadLine(conn_reader)
	weight_changes := ReadLine(conn_reader)

	threshold_battery := ReadLine(conn_reader)

	coordinator_evaluation_interval := ReadLine(conn_reader)

	for coordinator_changing.IsLocked() {
		time.Sleep(time.Millisecond * coordinator_transfer_time)
	}

	message := "vars changed\n" + interval_minutes + "\n" + interval_seconds + "\n" +
		weight_objects + "\n" + weight_runs + "\n" + weight_changes + "\n" + threshold_battery + "\n" + 
		coordinator_evaluation_interval + "\n"
	SendToAllDaemons(message)
	DialAndSend(coord_ip, coord_port, message)

	im, err := strconv.Atoi(interval_minutes)
	ErrorCheck(err, true)
	is, err := strconv.Atoi(interval_seconds)
	ErrorCheck(err, true)
	wo, err := strconv.Atoi(weight_objects)
	ErrorCheck(err, true)
	wr, err := strconv.Atoi(weight_runs)
	ErrorCheck(err, true)
	wc, err := strconv.Atoi(weight_changes)
	ErrorCheck(err, true)
	tb, err := strconv.Atoi(weight_objects)
	ErrorCheck(err, true)
	cei, err := strconv.Atoi(coordinator_evaluation_interval)
	ErrorCheck(err, true)

	reset_interval = time.Duration(im) * time.Minute + time.Duration(is) * time.Second
	objs_w = wo
	runs_w = wr
	changes_w = wc
	battery_treshold = tb
	coord_eval_interval = time.Duration(cei) * time.Second
	ToScreen("> Settings variables were changed")

	WriteLines(conn, "ack")
	return
}

func rc_VarsChanged(conn net.Conn, conn_reader *bufio.Reader) {
	interval_minutes := ReadIntLine(conn_reader)
	interval_seconds := ReadIntLine(conn_reader)

	weight_objects := ReadIntLine(conn_reader)
	weight_runs := ReadIntLine(conn_reader)
	weight_changes := ReadIntLine(conn_reader)

	threshold_battery := ReadIntLine(conn_reader)

	coordinator_evaluation_interval := ReadIntLine(conn_reader)

	reset_interval = time.Duration(interval_minutes) * time.Minute + time.Duration(interval_seconds) * time.Second
	objs_w = weight_objects
	runs_w = weight_runs
	changes_w = weight_changes
	battery_treshold = threshold_battery
	coord_eval_interval = time.Duration(coordinator_evaluation_interval) * time.Second

	ToScreen("> Settings variables were changed")
	WriteLines(conn, "ack")
	return
}
