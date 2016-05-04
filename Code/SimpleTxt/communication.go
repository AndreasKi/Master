//filename: communication.go
//information: created on 2th of November 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
)

var ip = "localhost" //The IP of this instance
var port = "8500"    //The port of this instance

//Handle connections
func listen() {
	listener, err := net.Listen("tcp", ":"+port)
	ErrorCheck(err, true)

	//Wait for GUI to initialize
	accepting.Lock()
	fmt.Println("ready")
	accepting.Unlock()
	for true {
		conn, err := listener.Accept()
		if ErrorCheck(err) {
			continue
		}
		defer conn.Close()
		go connection_handler(conn)
	}
}

//Serve connections
func connection_handler(conn net.Conn) {

	defer conn.Close()
	conn_reader := bufio.NewReader(conn)
	line := ReadLine(conn_reader)

	if line == "open" {
		rpc_open_file(conn, conn_reader)
	} else if line == "get" {
		rpc_get_text(conn, conn_reader)
	} else if line == "save" {
		rpc_save_file(conn, conn_reader)
	} else if line == "update" {
		rpc_update(conn, conn_reader)
	} else if line == "bold" {
		rpc_make_bold(conn, conn_reader)
	} else if line == "italic" {
		rpc_make_italic(conn, conn_reader)
	} else if line == "underlined" {
		rpc_make_underlined(conn, conn_reader)
	} else if line == "exit" {
		rpc_exit(conn, conn_reader)
	} else {
		WriteLines(conn, "Failed to understand message.")
		ErrorCheck(errors.New("Failed to understand message."))
	}
	return
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

//Reads a line from the connetion, and returns the string without newline char, after error checking
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
