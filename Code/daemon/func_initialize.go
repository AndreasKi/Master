//filename: func_initialize.go
//information: created on 2th of October 2015 by Andreas Kittilsland

package main

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"ioutil"
	"sync"
	"strings"
	"time"
)

func read_configuration() {
	//Read the config file
	bytes, err := ioutil.ReadFile("../config.cfg")
	ErrorCheck(err, true)
	cfg := string(bytes)

	lines := strings.Split(cfg, "\n")
	for _, line := range lines {
		if line != "" && len(line) > 2 && line[:2] != "//" {
			tuple := strings.Split(line, "=")
			variable := tuple[0]
			value := tuple[1]

			switch variable {
   			case "max_number_of_daemons":
   				target_size, err = strconv.Atoi(value)
   				ErrorCheck(true)
   			case "application_test_mode":
   				if value == "enabled" {
   					test_applications = true
   				}
   			case "coordinator_evaluation_interval":
   				interval, err := strconv.Atoi(value)
   				ErrorCheck(true)
   				coord_eval_interval = time.Duration(interval) * time.Second
   			case "file_weight":
   				objs_w, err = strconv.Atoi(value)
   				ErrorCheck(true)
   			case "application_runs_weight":
   				runs_w, err = strconv.Atoi(value)
   				ErrorCheck(true)
   			case "changes_weight":
   				changes_w, err = strconv.Atoi(value)
   				ErrorCheck(true)
      		case "battery_threshold":
   				battery_treshold, err = strconv.Atoi(value)
   				ErrorCheck(true)
   			default:
   				ErrorCheck(errors.New("config.cfg contained unknown variable " + variable), false)
   			}
		}
	}
}

//Look for a free port, and attempt to find a coordinator
func find_port_and_coordinator() {
	ip_self = "127.0.0.1"
	ip_to = "127.0.0.1"

	// Try to find a free port by brute force
	found_port := false
	for i := 0; i < target_size; i++ {
		port_int := 8590 + i
		port_i := strconv.Itoa(port_int)
		conn, err := net.Dial("tcp", ip_to+":"+port_i)
		if err != nil && !found_port {
			found_port = true
			port = port_i
			http_gui_port = strconv.Itoa((port_int - 8590) + 9000)
		} else if err == nil {
			WriteLines(conn, "joining")
			conn_reader := bufio.NewReader(conn)
			line := ReadLine(conn_reader) //communication.go
			if line == "coordinator" {    //We found the coordinator
				coord_ip = ip_to
				coord_port = port_i
			}
			conn.Close()
		}
	}
	if !found_port {
		ErrorCheck(errors.New("Failed to find a free port within range."), true)
	}
	if coord_port == "0" {
		if auto_coord {
			ChooseCoordinator()
		} else {
			BecomeCoordinator(true)
		}
	}

	is_running = make(map[int]bool) //for keeping track of which files ar opened
	go read_standard_input()        //communication.go
	return
}

//Set given port and become coordinator
func set_port_and_coordinator(suggested_port string) {
	port = suggested_port
	port_int, err := strconv.Atoi(port)
	ErrorCheck(err, true)
	http_gui_port = strconv.Itoa((port_int - 8590) + 9000)
	if auto_coord {
		ChooseCoordinator()
	} else {
		BecomeCoordinator(true)
	}

	is_running = make(map[int]bool) //for keeping track of which files ar opened

	return
}

//Set given port, and attempt to find a coordinator
func set_port_and_find_coordinator(suggested_port string) {
	//Looks for replies from potential coordinators
	ReadReply := func(conn net.Conn, d_port string, wg *sync.WaitGroup) {
		conn_reader := bufio.NewReader(conn)
		ch := make(chan string)
		err_chan := make(chan error)
		timeout := time.Tick(time.Second * 1)

		go func(ch chan string, err_chan chan error) {
			for {
				msg, err := conn_reader.ReadString('\n')
				if err != nil {
					err_chan <- err
					return
				}
				ch <- msg
			}
		}(ch, err_chan)
		for {
			select {
			case line := <-ch:
				if line == "coordinator\n" { //We found the coordinator
					coord_ip = ip_to
					coord_port = d_port
					wg.Done()
				}
				conn.Close()
				return
			case _ = <-err_chan:
				conn.Close()
				return
			case <-timeout:
				conn.Close()
				return
			}
		}
	}

	port = suggested_port
	coord_ip = ip_to
	coord_port = "0"
	port_int, err := strconv.Atoi(port)
	ErrorCheck(err, true)
	http_gui_port = strconv.Itoa((port_int - 8590) + 9000)

	var wg sync.WaitGroup
	wg.Add(1)
	for i := 0; i < target_size; i++ {
		ErrorCheck(err, true)
		d_port_int := 8590 + i
		d_port := strconv.Itoa(d_port_int)
		if d_port != suggested_port && coord_port == "0" {
			conn, err := net.Dial("tcp", ip_to+":"+d_port)
			if err == nil {
				WriteLines(conn, "joining")
				go ReadReply(conn, d_port, &wg)
			}
		}
	}
	wg.Wait()

	is_running = make(map[int]bool) //for keeping track of which files ar opened
	return
}

//Is this daemon synchronized with the coordinator after initial connection?
func Synced() bool {
	result := true
	if initialize.IsLocked() {
		result = false
	}

	return result
}

//Locks all the objects until we can reconnect to the coordinator and can receive a new coordinator version vector
func Connect() {
	//Lock the objects
	initialize.Lock()

	//Attempt to connect to coordinator
	nfound := true
	for nfound {
		ChooseCoordinator()
		if !is_coord {
			conn, err := Dial(coord_ip, coord_port)
			if ErrorCheck(err, false, false) {
				nfound = true
			} else {
				conn_reader := bufio.NewReader(conn)
				WriteLines(conn, "daemon req est conn", ip_to, port)
				line := ReadLine(conn_reader)
				if line == "N/A" {
					conn.Close()
					time.Sleep(time.Millisecond * 100)
					nfound = true
				} else if line == "ack" {
					coord_ver_vector_mu.Lock()
					coord_version_vector = ReadVersionVector(conn_reader)
					coord_ver_vector_mu.Unlock()
					conn.Close()
					nfound = false
				} else if line == "!coord" {
					coord_ip = ReadLine(conn_reader)
					coord_port = ReadLine(conn_reader)
					conn.Close()
					nfound = true
				}
			}
		} else {
			break
		}
	}

	initialize.Unlock()

	return
}
