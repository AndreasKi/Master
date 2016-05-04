package main

import (
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

//Get memory % of process with PID, and all child processes
func get_memory(pid string) float64 {
	//run "ps -p PID -o%cpu" in terminal to get information on CPU usage of this process
	cmd := exec.Command("ps", "-o", "pid,ppid,%mem")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		ErrorCheck(errors.New("ps -o pid,ppid,%mem returned with error: "+err.Error()), true, true, true)
	}

	//Extract the percentage
	res := 0.0
	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Split(line, " ")
		var t_line []string
		for _, t := range tokens {
			if t != "" && t != "\t" {
				t_line = append(t_line, t)
			}
		}
		if t_line[0] == pid || t_line[1] == pid {
			float, err := strconv.ParseFloat(t_line[2][:len(t_line[2])-1], 64)
			ErrorCheck(err, true)
			res = res + float
		}
	}

	return res
}

//Get cpu % of process with PID, and all child processes
func get_cpu(pid string) float64 {
	//run "ps -p PID -o%cpu" in terminal to get information on CPU usage of this process
	cmd := exec.Command("ps", "-o", "pid,ppid,%cpu")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		ErrorCheck(errors.New("ps -o pid,ppid,%cpu returned with error: "+err.Error()), true, true, true)
	}

	//Extract the percentage
	res := 0.0
	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Split(line, " ")
		var t_line []string
		for _, t := range tokens {
			if t != "" && t != "\t" {
				t_line = append(t_line, t)
			}
		}
		if t_line[0] == pid || t_line[1] == pid {
			float, err := strconv.ParseFloat(t_line[2][:len(t_line[2])-1], 64)
			ErrorCheck(err, true)
			res = res + float
		}
	}

	return res
}
