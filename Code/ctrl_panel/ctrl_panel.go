//filename: ctrl_panel.go
//information: created on 30th of August 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"errors"
	"os/exec"
	"io/ioutil"
	"strings"
	"runtime"
	"strconv"
	"time"
)

func main() {
	//Get PID
	pid_string = strconv.Itoa(os.Getpid())

	read_configuration()

	check_connectivity()

	go open_browser()

	run_gui()
}

func open_browser() {
	time.Sleep(time.Second)

	//Open GUI in browser
	if runtime.GOOS == "darwin" {
		//OSX
		exec.Command("open", "http://localhost"+http_gui_port).Start()
	} else {
		//Display Wall
		exec.Command("chromium-browser", "--user-data-dir=/tmp", "http://localhost"+http_gui_port).Start()
	}

	return
}

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
   				ErrorCheck(err,true)
   			case "application_test_mode":
   			case "coordinator_evaluation_interval":
   			case "file_weight":
   			case "application_runs_weight":
   			case "changes_weight":
      		case "battery_threshold":
   			default:
   				ErrorCheck(errors.New("config.cfg contained unknown variable " + variable), false)
   			}
		}
	}
}

func check_connectivity() {
	for i := 0; i < target_size; i++ {
		port_int := 8590 + i
		port_i := strconv.Itoa(port_int)
		conn, err := net.Dial("tcp", "127.0.0.1:"+port_i)
		if err == nil {
			fmt.Fprintf(conn, "test conn\n")
			conn_reader := bufio.NewReader(conn)
			line, err := conn_reader.ReadString('\n')
			ErrorCheck(err, true)
			if line == "accepted\n" {
				d_pid, err := conn_reader.ReadString('\n')
				ErrorCheck(err, true)
				d_pid = d_pid[:len(d_pid)-1]

				found := false
				for _, d := range daemons {
					if d.port == port_i {
						found = true
						break
					}
				}
				if !found {
					new_daemon := daemon{"127.0.0.1", port_i, d_pid}
					daemons = append(daemons, new_daemon)
					AddNotification("New daemon was found at 127.0.0.1:" + port_i)
				}
			}
			conn.Close()
		} else {
			for i, d := range daemons {
				if d.port == port_i {
					daemons[i] = daemons[len(daemons)-1]
					daemons = daemons[:len(daemons)-1]
					break
				}
			}
		}
	}
}

func read_daemon_output(scanner *bufio.Scanner, ip_d string, port_d string) {
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			fmt.Println("Daemon "+ip_d+":"+port_d+": ", line)
		}
	}

	daemons_mu.Lock()
	for index, cur_daemon := range daemons {
		if cur_daemon.ip == ip_d && cur_daemon.port == port_d {
			daemons[index] = daemons[len(daemons)-1]
			daemons = daemons[:len(daemons)-1]
			break
		}
	}
	daemons_mu.Unlock()
	fmt.Println("Lost connection to daemon at " + ip_d + ":" + port_d)
	AddNotification("Lost connection to daemon at " + ip_d + ":" + port_d)
}
