package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"
)

func add(target string, times int, adds_per_refresh int) {
	refresh(target[7:])
	time.Sleep(time.Second * 5)
	iteration := 0
	refresh_counter := 0
	
	dat, err := ioutil.ReadFile("testfile")
    ErrorCheck(err, true)
    test_content := string(dat)

	for !stop {
		fname := strconv.Itoa(iteration) + ".txt"
		f, err := os.Create("../files/" + target + "/" + fname)
		ErrorCheck(err, true)
		_, err = f.WriteString(test_content)
		ErrorCheck(err, true)
		f.Close()
		conn, err := net.Dial("tcp", "127.0.0.1:"+target[7:])
		ErrorCheck(err, true)

		fmt.Fprintf(conn, "new file\n"+fname+"\n")
		conn_reader := bufio.NewReader(conn)
		line, err := conn_reader.ReadString('\n')
		if line != "ack\n" {
			ErrorCheck(errors.New("Failed to comprehend reply from daemon: "+line), true)
		}
		conn.Close()

		iteration++
		fmt.Println("Added " + fname)
		if times != -1 {
			if !(iteration < times) {
				stop = true
			}
		}

		if adds_per_refresh != -1 {
			refresh_counter++
		}
		
		if refresh_counter >= adds_per_refresh {
			refresh(target[7:])
			refresh_counter = 0
			time.Sleep(time.Second*5)
		}
	}
	os.Exit(1)
}

func refresh(port string) {
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	ErrorCheck(err, true)

	fmt.Fprintf(conn, "refresh\n"+mode+"\n")
	conn_reader := bufio.NewReader(conn)
	line, err := conn_reader.ReadString('\n')
	if line != "ack\n" {
		ErrorCheck(errors.New("Failed to comprehend reply from daemon: "+line), true)
	}
	conn.Close()
}
