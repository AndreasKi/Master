//filename: func_coordinator.go
//information: created on 8th of October 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

var coord_eval_interval = time.Second * 5

var objs_w = 1            //Weight for each object towards coordination election
var runs_w = 10           //Weight for application run towards coordination election
var changes_w = 5         //Weight for each local change towards coordination election
var battery_treshold = 20 //Lowest battery value allowed for coordinator

const coordinator_transfer_time = 500 //How long the system has to transfer the coordinator to a different daemon before changes start coming in agian.

var coordinator_changing = ExpandedLock{}
var coordinator_serving_atm = 0 //How many daemons is the coordinator serving right now

var is_coord = false       //Whether or not this instance is the coordinator
var coord_ip = "127.0.0.1" //The IP of the coordinator
var coord_port = "0"       //The port of the coordinator

var current_updates = sync.WaitGroup{} //Counts number of updates that are curently being pushed

func CalcCoordScore(objs, runs, changes, battery_percentage int, is_charging bool) int {
	totScore := 0

	battScore := 0
	if is_charging {
		battScore = 100
	} else {
		battScore = battery_percentage
	}

	if battScore > battery_treshold {
		totScore = battery_treshold + (objs*objs_w + changes*changes_w + runs*runs_w)
	} else {
		totScore = battery_percentage
	}
	return totScore
}

func ChooseCoordinator() {
	//See if a new daemon should be chosen to be coordinator
	if coordinator_changing.IsLocked() {
		//Dont look for a new coordinator if havent even finished changing after previous election.
		return
	}

	//Make sure we discard expired values before calculating
	ResetMetrics()

	//Get battery status
	isCharging, battPercent := CheckBattery()

	//Initialize to self
	high_port := port
	highest := CalcCoordScore(nr_local_objs, nr_of_runs, nr_of_changes, battPercent, isCharging)

	//Request values from other daemons
	for index, c_daemon := range daemons {
		conn, err := Dial(c_daemon.ip, c_daemon.port)
		if !ErrorCheck(err) {
			WriteLines(conn, "request state")

			conn_reader := bufio.NewReader(conn)
			line := ReadLine(conn_reader)

			if line == "accepted" {
				nr_objs := ReadIntLine(conn_reader)
				runs := ReadIntLine(conn_reader)
				changes := ReadIntLine(conn_reader)
				curr_d_isCharging_s := ReadLine(conn_reader)
				curr_d_battPercent := ReadIntLine(conn_reader)

				curr_d_isCharging := false
				if curr_d_isCharging_s == "true" {
					curr_d_isCharging = true
				}

				score := CalcCoordScore(nr_objs, runs, changes, curr_d_battPercent, curr_d_isCharging)

				//The daemon was found to have a higher score than our current largest daemon
				if highest < score {
					highest = score
					high_port = c_daemon.port
				}
			} else {
				conn.Close()
				ErrorCheck(errors.New("Unexpected reply from daemon "+c_daemon.ip+":"+c_daemon.port+"; "+line), true)
			}
			conn.Close()
		} else {
			RemoveDaemon(index)
		}
	}

	if !is_coord && coord_port != "0" {
		//Compare to current coordinator
		conn_coord, err := Dial(coord_ip, coord_port)
		if !ErrorCheck(err) {
			WriteLines(conn_coord, "request state")

			conn_reader := bufio.NewReader(conn_coord)
			line := ReadLine(conn_reader)

			if line == "accepted" {
				nr_objs := ReadIntLine(conn_reader)
				runs := ReadIntLine(conn_reader)
				changes := ReadIntLine(conn_reader)
				curr_d_isCharging_s := ReadLine(conn_reader)
				curr_d_battPercent := ReadIntLine(conn_reader)

				curr_d_isCharging := false
				if curr_d_isCharging_s == "true" {
					curr_d_isCharging = true
				}

				score := CalcCoordScore(nr_objs, runs, changes, curr_d_battPercent, curr_d_isCharging)

				//The coordinator was found to have a higher or equal score than/as our current highest scoring daemon
				if highest <= score {
					highest = score
					high_port = coord_port
				}
			} else {
				conn_coord.Close()
				ErrorCheck(errors.New("Unexpected reply from daemon "+coord_ip+":"+coord_port+"; "+line), true)
			}
			conn_coord.Close()
		}
	}

	if high_port == port && coord_port != port {
		if coord_port == "0" {
			BecomeCoordinator(true)
		} else {
			BecomeCoordinator(false)
		}
		go continous_coordinator_evaluation()
	} //If we didnt become the coordinator, wait for the new coordinator to claim its position
	return
}

//An update was made, push it
func coord_InstigateUpdate(id string, version string, attributes map[string]string, timer string, obj_type string) error {
	coordinator_serving_atm++
	current_updates.Add(1)
	id_int, err := strconv.Atoi(id)
	ErrorCheck(err, true)
	version_int, err := strconv.Atoi(version)
	ErrorCheck(err, true)
	//id_checked is the same as id_int if it is not a new object. However it contains a proper id if it is a new object
	id_checked, err := AddObject(id_int, version_int, attributes, timer, obj_type) //mutators.go
	if err != nil {
		coordinator_serving_atm--
		return err
	}
	if _, found := locks[id_checked]; !found { //If we made a new object, lock it
		locks[id_checked] = &ExpandedLock{} //lock.go
		locks[id_checked].Lock()
	}

	//Send the update to all daemons in the network
	message_string := "update\n" + strconv.Itoa(id_checked) + "\n" + version + "\n" + obj_type + "\n" + timer + "\n"
	for index, item := range attributes {
		message_string = message_string + index + " " + item + "\n"
	}
	message_string = message_string + "done\n"

	SendToAllDaemons(message_string) //communication.go

	locks[id_checked].Unlock() //lock.go
	current_updates.Done()
	coordinator_serving_atm--
	return nil
}

//An update was made, push it
func coord_InstigateDelete(id string, timer string) error {
	coordinator_serving_atm++
	current_updates.Add(1)
	id_int, err := strconv.Atoi(id)
	ErrorCheck(err, true)

	//Lock object, output error if object lock is non-existant
	if _, ok := locks[id_int]; ok {
		locks[id_int].Lock()
	} else {
		ErrorCheck(errors.New("Attempted to lock non-existant object"), true)
	}
	//Do the remove operation, and return error if it fails
	if err = RemoveObject(id_int, timer); err != nil { //mutators.go
		locks[id_int].Unlock() //lock.go
		coordinator_serving_atm--
		return err
	}

	//Send the update to all daemons in the network
	SendToAllDaemons("delete\n" + id + "\n" + timer + "\n") //communication.go

	current_updates.Done()
	coordinator_serving_atm--
	return nil
}

func coord_DaemonEstablishConnection(conn net.Conn, conn_reader *bufio.Reader) {
	coordinator_serving_atm++
	refreshing.Lock()
	//Another daemon wishes to join,
	//get its details
	ip_d := ReadLine(conn_reader)
	port_d := ReadLine(conn_reader)

	ver_vector_mu.ReadOnly()
	WriteVersionVector(conn, version_vector)
	ver_vector_mu.Editable()

	//Add the daemon to our local list
	is_new := AddDaemon(ip_d, port_d) //mutators.go

	if is_new {
		//Inform the other daemons that a new daemon joined
		SendToAllDaemons("daemon est conn\n" + ip_d + "\n" + port_d + "\n")

		//Send addresses of all the other daemons to the new daemon
		daemons_mu.Lock()
		for _, d := range daemons {
			err := DialAndSend(ip_d, port_d, "daemon est conn\n"+d.ip+"\n"+d.port+"\n")
			ErrorCheck(err, true)
		}
		daemons_mu.Unlock()
	}

	new_daemon := daemon{ip_d, port_d}
	sendApplicationsTo_mu.Lock()
	sendApplicationsTo = append(sendApplicationsTo, new_daemon)
	sendApplicationsTo_mu.Unlock()

	refreshing.Unlock()
	coordinator_serving_atm--
	return
}

//Someone is requesting access to a remote file. Find it.
func coord_request_file_access(conn net.Conn, conn_reader *bufio.Reader, id int) {
	coordinator_serving_atm++
	success := false
	daemons_mu.Lock()
	for _, dmn := range daemons {
		//Ask for the file
		file_conn, err := Dial(dmn.ip, dmn.port)
		ErrorCheck(err, true)

		WriteLines(file_conn, "req file access")
		WriteIntLine(file_conn, id)

		file_reader := bufio.NewReader(file_conn)
		outcome := ReadLine(file_reader)
		if outcome == "ok" {
			url := ReadLine(file_reader)
			WriteLines(conn, outcome, url)
			success = true
			file_conn.Close()
			break
		} else if outcome == "sync conflict" {
			file_conn.Close()
			break
		}
		file_conn.Close()
	}
	if !success {
		WriteLines(conn, "nf")
	}
	daemons_mu.Unlock()
	coordinator_serving_atm--
	return
}

func coord_update_app_list(list []application) {
	coordinator_serving_atm++
	message := "app list\n"
	for _, app := range list {
		message = message + app.app_name + "\n"
	}
	message = message + "done\n"
	SendToAllDaemons(message)
	coordinator_serving_atm--
	return
}

func continous_coordinator_evaluation() {
	time.Sleep(coord_eval_interval)
	for is_coord {
		AutomaticCoordinator()
		time.Sleep(coord_eval_interval)
	}
}

func coord_finish_tasks(new_coord_ip string, new_coord_port string) {
	for true {
		if coordinator_serving_atm == 0 {
			break
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
	is_coord = false
	AddDaemon(ip_to, port)
	ToScreen(daemons)

	coord_ip = new_coord_ip
	coord_port = new_coord_port

	daemons_mu.Lock()
	for i, d := range daemons {
		if d.ip == coord_ip && d.port == coord_port {
			RemoveDaemon(i)
			break
		}
	}
	daemons_mu.Unlock()

	DialAndSend(coord_ip, coord_port, "coord tasks done\n")
	SendToAllDaemons("coordinator changed\n")

	coordinator_changing.Unlock()
	ToScreen("> I am a profane daemon")
	ToScreen("> Status: Has " + strconv.Itoa(len(daemons)+2) + " daemons in network.")
	return
}

//Finds and adds all applications on this device
func coord_FindApplications() {
	applications_mu.Lock()
	local_apps_mu.Lock()

	changed, previous_list, new_apps := RetrieveLocalApplications()

	//Add new apps to the global list
	for _, app := range new_apps {
		if i, ok := AppIn_Struct(applications, app); !ok {
			applications = append(applications, application{app, 1})
		} else {
			applications[i].count++
		}
	}

	//Find difference between current and previous list to find removed apps
	var removed_apps []string
	for _, app := range previous_list {
		if _, ok := AppIn_String(local_apps, app); !ok {
			removed_apps = append(removed_apps, app)
		}
	}

	//Remove apps that are no longer in the local list
	if len(removed_apps) > 0 {
		changed = true
		for _, app := range removed_apps {
			fmt.Println("removed app:")
			fmt.Println(app)
			if i, ok := AppIn_Struct(applications, app); ok {
				applications[i].count--
				if applications[i].count < 1 {
					applications[i] = applications[len(applications)-1]
					applications = applications[:len(applications)-1]
				}
			}
		}
		for _, app := range removed_apps {
			if i, ok := AppIn_String(local_apps, app); ok {
				local_apps[i] = local_apps[len(local_apps)-1]
				local_apps = local_apps[:len(local_apps)-1]
			}
		}
	}
	if changed {
		go coord_update_app_list(applications)
	}
	applications_mu.Unlock()
	local_apps_mu.Unlock()
}
