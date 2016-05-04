package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func main() {
	read_configuration()

	if len(os.Args) == 2 {
		secs, err := strconv.Atoi(os.Args[1])
		ErrorCheck(err, true)

		log_time = secs
		fmt.Println("Running logging for " + strconv.Itoa(log_time) + "s...")
	} else {
		fmt.Println("Running logging until a received <stop> command...")
	}

	start = time.Now()
	FindDaemons()
	go monitor_performance()
	ReadSTDIN()

	finished.Wait()

	return
}

func FindDaemons() {
	daemons_mu.Lock()
	for i := 0; i < target_size; i++ {
		port_int := 8590 + i
		port_i := strconv.Itoa(port_int)
		conn, err := net.Dial("tcp", "127.0.0.1:"+port_i)
		if err == nil {
			fmt.Fprintf(conn, "test conn\n")
			conn_reader := bufio.NewReader(conn)
			line, err := conn_reader.ReadString('\n')
			if err != nil {
				for index, d := range daemons {
					if d.port == port_i {
						daemons[index].dead = true
						break
					}
				}
			} else {
				if line == "accepted\n" {
					d_pid, err := conn_reader.ReadString('\n')
					if err != nil {
						for index, d := range daemons {
							if d.port == port_i {
								daemons[index].dead = true
								break
							}
						}
					} else {
						d_pid = d_pid[:len(d_pid)-1]

						found := false
						for _, d := range daemons {
							if d.port == port_i {
								found = true
								break
							}
						}
						if !found {
							new_daemon := daemon{ip: "127.0.0.1", port: port_i, pid: d_pid, dead: false, start_time: runs_since_start}
							daemons = append(daemons, new_daemon)
						}
					}
				}
			}
			conn.Close()
		} else {
			for index, d := range daemons {
				if d.port == port_i {
					daemons[index].dead = true
					break
				}
			}
		}
	}
	daemons_mu.Unlock()
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

func ReadSTDIN() {
	std_reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter cmd: ")
	for !stop {
		line, _, err := std_reader.ReadLine()
		ErrorCheck(err, true)

		cmd := string(line)
		if cmd == "stop" {
			stop = true
			Output()
			break
		}
	}
	return
}
