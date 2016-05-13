//filename: support_communication.go
//information: created on 22th of September 2015 by Andreas Kittilsland
//Contains general communcation functions
package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Overridden if no port is given on run. Set to localhost.

var ip_self = "192.168.1.138" //The IP of this instance
var ip_to = "192.168.1.138"   //The IP of this instance

//var ip_to = "192.168.1.86"    //remote IP redirected back to self

var port = "8590" //The port of this instance if no port is given on run. Changed if other daemons are found running already.

//Handle connections
func listen() {
	listener, err := net.Listen("tcp", ":"+port)
	//listener, err := net.Listen("tcp", ":"+port)
	ErrorCheck(err, true)

	for true {
		conn, err := listener.Accept()
		if ErrorCheck(err) {
			continue
		}
		defer conn.Close()
		//rc_* functions are in remote_commands.go
		go rc_ConnectionHandler(conn)
	}
	return
}

func read_standard_input() {
	std_reader := bufio.NewReader(os.Stdin)
	for true {
		line, _, err := std_reader.ReadLine()
		if err != nil && err.Error() != "EOF" {
			ErrorCheck(err, true)
		} else if err == nil {
			cmd := string(line)
			ToScreen(cmd)
			if cmd == "set coordinator" {
				if coord_port == "0" {
					BecomeCoordinator(true)
				} else {
					BecomeCoordinator(false)
				}
			} else if cmd == "automatic coordinator" {
				AutomaticCoordinator()
			}
		}
	}
	return
}

//Writes to stdin of the change tracker
func ToChangeTracker(msg string) {
	ChangeDetector_InPipe.Write([]byte(msg))
	return
}

//Reads a line from the connection, and returns the string without newline char, after error checking
func ReadDocumentFragment(conn_reader *bufio.Reader) string {
	text := ""
	for true {
		line, err := conn_reader.ReadString('\n')
		ErrorCheck(err, true, true, true)

		if line != "<--done-->\n" {
			text = text + line
		} else {
			break
		}
	}

	return text[:len(text)-1]
}

//Reads a line from the connetion, and returns the string without newline char, after error checking
func ReadLine(conn_reader *bufio.Reader) string {
	line, err := conn_reader.ReadString('\n')
	ErrorCheck(err, true, true, true)

	return line[:len(line)-1]
}

//Writes lines to the connection
func WriteLines(conn net.Conn, lines ...string) {
	for _, line := range lines {
		fmt.Fprintf(conn, line+"\n")
	}
	return
}

//Reads a line from the connection, and returns the integer found in the line, after error checking
func ReadIntLine(conn_reader *bufio.Reader) int {
	line, err := conn_reader.ReadString('\n')
	ErrorCheck(err, true, true, true)

	line_int, err := strconv.Atoi(line[:len(line)-1])
	ErrorCheck(err, true, true, true)

	return line_int
}

//Writes an integer to the connection
func WriteIntLine(conn net.Conn, integer int) {
	fmt.Fprintf(conn, strconv.Itoa(integer)+"\n")
	return
}

//Read an entire map
func ReadAttributesMap(conn_reader *bufio.Reader) map[string]string {
	attributes := make(map[string]string)
	for true {
		tuple_line, err := conn_reader.ReadString('\n')
		ErrorCheck(err, true, true, true)

		if tuple_line != "done\n" {
			tuple := strings.Split(tuple_line, " ")
			for i := 1; i < len(tuple); i++ { //Do this incase there are whitespaces in the content (f. ex. name)
				attributes[tuple[0]] = attributes[tuple[0]] + " " + tuple[i]
			}
			attributes[tuple[0]] = strings.TrimSpace(attributes[tuple[0]]) //Remove leading space
		} else {
			break
		}
	}
	return attributes
}

//Write an entire map
func WriteAttributesMap(conn net.Conn, attributes map[string]string) {
	for index, item := range attributes {
		fmt.Fprintf(conn, index+" "+item+"\n")
	}
	fmt.Fprintf(conn, "done\n")
	return
}

//Read a slice from reader
func ReadSlice(conn_reader *bufio.Reader) []string {
	var slice []string
	for true {
		line, err := conn_reader.ReadString('\n')
		ErrorCheck(err, true, true, true)

		if line != "done\n" {
			slice = append(slice, line[:len(line)-1])
		} else {
			break
		}
	}

	return slice
}

//Write slice to connection
func WriteSlice(conn net.Conn, slice []string) {
	for _, item := range slice {
		fmt.Fprintf(conn, item+"\n")
	}
	fmt.Fprintf(conn, "done\n")
	return
}

//Read a slice from reader
func ReadApps(conn_reader *bufio.Reader) []application {
	var slice []application
	for true {
		line, err := conn_reader.ReadString('\n')
		ErrorCheck(err, true, true, true)

		if line != "done\n" {
			slice = append(slice, application{line[:len(line)-1], 1})
		} else {
			break
		}
	}

	return slice
}

//Write slice to connection
func WriteApps(conn net.Conn, slice []application) {
	for _, item := range slice {
		fmt.Fprintf(conn, item.app_name+"\n")
	}
	fmt.Fprintf(conn, "done\n")
	return
}

//Write the time now to connection
func WriteTimeNow(conn net.Conn) {
	timer := time.Now().Format(time.RFC3339)
	fmt.Fprintf(conn, timer)
	fmt.Fprintf(conn, "\n")
	return
}

//Write a duration to the connection
func WriteDuration(conn net.Conn, dur time.Duration) {
	fmt.Fprintf(conn, dur.String())
	fmt.Fprintf(conn, "\n")
	return
}

//Send a message to the given address
func DialAndSend(ip_d string, port_d string, message string) error {
	var outcome error
	outcome = nil

	conn, err := Dial(ip_d, port_d)
	if err != nil {
		outcome = err
	} else {
		fmt.Fprintf(conn, message)
		reader := bufio.NewReader(conn)
		status, err := reader.ReadString('\n')
		ErrorCheck(err, true, true, true)
		status = status[:len(status)-1]
		if status != "ack" {
			outcome = errors.New(status)
		}
		conn.Close()
	}

	return outcome
}

func WriteVersionVector(conn net.Conn, vector map[int]int) {
	for id, version := range vector {
		fmt.Fprintf(conn, strconv.Itoa(id)+"-"+strconv.Itoa(version)+"\n")
	}
	fmt.Fprintf(conn, "done\n")
	return
}

func ReadVersionVector(conn_reader *bufio.Reader) map[int]int {
	vector := make(map[int]int)
	for true {
		line, err := conn_reader.ReadString('\n')
		ErrorCheck(err, true, true, true)

		if line != "done\n" {
			formated_line := strings.Split(line[:len(line)-1], "-")
			id, err := strconv.Atoi(formated_line[0])
			ErrorCheck(err, true, true, true)
			version, err := strconv.Atoi(formated_line[1])
			ErrorCheck(err, true, true, true)
			vector[id] = version
		} else {
			break
		}
	}

	return vector
}

//Send all objects in the passed vector to the specified daemon
func SendObjsInVector(vector map[int]int, ip_d string, port_d string) {
	obj_mu.ReadOnly()
	for index, version := range vector {
		if version > objects[index].version {
			ErrorCheck(errors.New("Syncronization error: daemon has newer version of object than coordinator"), true)
		} else {
			conn, err := Dial(ip_d, port_d)
			ErrorCheck(err, true)
			fmt.Fprintf(conn, "update\n"+strconv.Itoa(index)+"\n"+strconv.Itoa(objects[index].version)+"\n"+objects[index].obj_type+"\n")
			timer := time.Now().Format(time.RFC3339)
			fmt.Fprintf(conn, timer)
			fmt.Fprintf(conn, "\n")
			for i, item := range objects[index].attributes {
				fmt.Fprintf(conn, i+" "+item+"\n")
			}
			fmt.Fprintf(conn, "done\n")
			conn.Close()
		}
	}
	obj_mu.Editable()
	return
}

//Send all objects in the passed vector to the specified daemon
func RequestObjsInVector(vector map[int]int) {
	conn, err := Dial(coord_ip, coord_port)
	ErrorCheck(err, true)
	fmt.Fprintf(conn, "req objs\n")
	fmt.Fprintf(conn, ip_to+"\n")
	fmt.Fprintf(conn, port+"\n")
	WriteVersionVector(conn, vector)
	conn_reader := bufio.NewReader(conn)
	line := ReadLine(conn_reader)
	if line != "ack" {
		ErrorCheck(err, true, true, true)
	}
	conn.Close()

	return
}

//Send a message to all the known daemons
func SendToAllDaemons(message string) {
	//Function definition for all the go-routines we make, one for each daemon
	var wg sync.WaitGroup
	send_message := func(ip_d string, port_d string, index int) {
		conn, err := Dial(ip_d, port_d)
		if !ErrorCheck(err, true, true, true) {
			fmt.Fprintf(conn, message)
			reader := bufio.NewReader(conn)
			line, err := reader.ReadString('\n')
			ErrorCheck(err, true, true, true)
			status := line[:len(line)-1]
			if status != "ack" {
				ErrorCheck(err, true, true, true)
			}
			conn.Close()
		}
		wg.Done()
		return
	}

	//Spawn go-routines
	daemons_mu.ReadOnly()
	for index, daemon_instc := range daemons {
		wg.Add(1)
		go send_message(daemon_instc.ip, daemon_instc.port, index)
	}
	daemons_mu.Editable()
	wg.Wait()
	return
}

func Dial(ip_dest string, port_dest string) (net.Conn, error) {
	var port_to_dial string
	if emulate_network {
		em_port_int, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		em_port_int = em_port_int - 1000
		port_to_dial = strconv.Itoa(em_port_int)
	} else {
		port_to_dial = port_dest
	}

	conn, err := net.Dial("tcp", ip_dest+":"+port_to_dial)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func ToScreen(msg ...interface{}) {
		//fmt.Println(ip_to+":"+port+" - ", msg)
	
	if is_coord {
		fmt.Println(ip_to+":"+port+" - ", msg)
	} else if ip_to == "127.0.0.1" {
		fmt.Println(ip_to+":"+port+" - ", msg)
	
	}
	return
}

func Dump(msg ...interface{}) {
	fmt.Println(msg)
	return
}