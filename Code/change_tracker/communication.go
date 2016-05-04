//filename: communication.go
//information: created on 26th of November 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetPID() {
	fmt.Println("PID:" + strconv.Itoa(os.Getpid()))
}

//Notifies daemon of changed file
func ChangeFile(output string) {
	fmt.Println(output)
}

//Notifies daemon of added file
func AddFile(item string) {
	fmt.Println("add:" + item)
}

//Notifies daemon of removed files
func DeleteFile(item string) {
	fmt.Println("delete:" + item)
	fmt.Println("!delete:" + item)
}

//Notifies daemon of a name change
func NameChangeFile(from string, to string) {
	fmt.Println(from + ":" + to)
}

func FinishSyncedScan() {
	fmt.Println("sync run done\n")
}

//Read commands from daemon
func ReadStdin() {
	std_reader := bufio.NewReader(os.Stdin)
	for true {
		line, _, err := std_reader.ReadLine()
		ErrorCheck(err, true)

		cmd := string(line)
		if cmd == "run" {
			DoWalk(true)
		} else if cmd == "sync run" {
			DoWalk(false)
		} else if cmd == "edit" {
			line, _, err = std_reader.ReadLine()
			ErrorCheck(err, true)
			cmd = string(line)
			ChangeFile(cmd)
		} else if cmd == "add" {
			line, _, err = std_reader.ReadLine()
			ErrorCheck(err, true)
			cmd = string(line)
			AddFile(cmd)
		} else if cmd == "name_change" {
			line, _, err = std_reader.ReadLine()
			ErrorCheck(err, true)
			cmd = string(line)
			t := strings.Split(cmd, ":")
			NameChangeFile(t[0], t[1])
		}
	}
}
