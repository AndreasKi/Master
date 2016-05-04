package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
)

var stop = false
var mode = "objs"

func main() {
	var target string
	var ref_rate int
	nf := -1
	var err error
	if len(os.Args) == 4 {
		nf, err = strconv.Atoi(os.Args[3])
		ErrorCheck(err, true)
		target = os.Args[1]
		ref_rate, err = strconv.Atoi(os.Args[2])
		ErrorCheck(err, true)
		fmt.Println("Adding " + strconv.Itoa(nf) + " files")
	} else if len(os.Args) == 3 {
		target = os.Args[1]
		ref_rate, err = strconv.Atoi(os.Args[2])
		ErrorCheck(err, true)
		fmt.Println("Adding files until until a received <stop> command...")
	} else {
		ErrorCheck(errors.New("Received no arguments! need target daemon, refresh frequency, and optionally number of seconds to run."), true)
	}

	go add(target, nf, ref_rate)
	ReadSTDIN()

	return
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
			fmt.Println("Stopping")
			break
		} else if cmd == "apps"{
			mode = cmd
			fmt.Println("Refresh mode set to apps")
		} else if cmd == "objs" {
			mode = cmd
			fmt.Println("Refresh mode set to objs")
		}
	}
	return
}
