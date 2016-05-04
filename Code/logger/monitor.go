package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var cpu_history []float64
var mem_history []int
var num_daemons []int
var runs_since_start = 0

type process_information struct {
	pid  string
	ppid string
	cpu  float64
	mem  int
}

type children struct {
	pids []string
}

func monitor_performance() {
	for !stop {
		start_time := time.Now()
		alive_daemons := 0
		tot_mem := 0
		tot_cpu := 0.0
		wg := sync.WaitGroup{}

		var err error
		var pids map[string]process_information
		var child_pids map[string]children
		if len(daemons) > 0 {
			pids, child_pids, err = get_stats()
			if err != nil {
				Output()
				os.Exit(1)
			}
		}

		daemons_mu.Lock()
		for i, d := range daemons {
			if !d.dead {
				alive_daemons++
				if d.pid != "0" {
					wg.Add(1)
					go func(index int) {
						not_found := true
						cpu_val := 0.0
						mem_val := 0
						if PI, found := pids[daemons[index].pid]; found {
							not_found = false
							cpu_val = cpu_val + PI.cpu
							mem_val = mem_val + PI.mem
						}
						if these_pids, found := child_pids[daemons[index].pid]; found {
							not_found = false
							for _, this_pid := range these_pids.pids {
								cpu_val = cpu_val + pids[this_pid].cpu
								mem_val = mem_val + pids[this_pid].mem
							}
						}
						if not_found {
							daemons[index].cpu_history = append(daemons[index].cpu_history, -1.0)
							daemons[index].dead = true

							daemons[index].mem_history = append(daemons[index].mem_history, -1)
							daemons[index].dead = true
						} else {
							daemons[index].cpu_history = append(daemons[index].cpu_history, cpu_val)
							tot_cpu = tot_cpu + cpu_val

							daemons[index].mem_history = append(daemons[index].mem_history, mem_val)
							tot_mem = tot_mem + mem_val
						}
						wg.Done()
					}(i)
				} else {
					d.mem_history = append(d.mem_history, -1.0)
					d.cpu_history = append(d.cpu_history, -1.0)
				}
			} else {
				d.mem_history = append(d.mem_history, -1.0)
				d.cpu_history = append(d.cpu_history, -1.0)
			}
		}
		wg.Wait()
		daemons_mu.Unlock()

		cpu_history = append(cpu_history, tot_cpu)
		mem_history = append(mem_history, tot_mem)
		num_daemons = append(num_daemons, alive_daemons)
		runs_since_start++

		if log_time != -1 && log_time <= runs_since_start*interval {
			stop = true
			Output()
			break
		} else {
			FindDaemons()

			wait_time := time.Second*time.Duration(interval) - time.Since(start_time)
			if wait_time > 0 {
				time.Sleep(wait_time)
			} else {
				ErrorCheck(errors.New("Monitor loop took longer than "+strconv.Itoa(interval)+"s. Logging delayed..."), false)
			}
			fmt.Println("Seconds: " + strconv.Itoa(log_time-runs_since_start*interval))
		}

	}
	finished.Wait()
	os.Exit(1)
}

func get_stats() (map[string]process_information, map[string]children, error) {
	cmd := exec.Command("top", "-stats", "pid,ppid,cpu,mem", "-l", strconv.Itoa(interval-1))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, nil, err
	}

	var pid_map map[string]process_information
	pid_map = make(map[string]process_information)

	var children_map map[string]children
	children_map = make(map[string]children)

	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		t := strings.Split(line, " ")
		var tokens []string
		for _, token := range t {
			if token != "" && token != "\t" && token != "\n" {
				tokens = append(tokens, token)
			}
		}
		if len(tokens) == 0 {
			continue
		} else if _, err := strconv.Atoi(tokens[0]); err == nil && len(tokens) == 4 {
			pid := tokens[0]
			ppid := tokens[1]
			cpuPerc_i := tokens[2]
			memStr_i := tokens[3]
			if memStr_i[len(memStr_i)-1:len(memStr_i)] == "\n" {
				memStr_i = memStr_i[:len(memStr_i)-1]
			}
			if memStr_i[len(memStr_i)-1:len(memStr_i)] == "+" {
				memStr_i = memStr_i[:len(memStr_i)-1]
			}
			if memStr_i[len(memStr_i)-1:len(memStr_i)] == "-" {
				memStr_i = memStr_i[:len(memStr_i)-1]
			}

			cpu, err := strconv.ParseFloat(cpuPerc_i, 64)
			ErrorCheck(err, true)

			mem, err := strconv.Atoi(memStr_i[:len(memStr_i)-1])
			ErrorCheck(err, true)
			if memStr_i[len(memStr_i)-1:len(memStr_i)] == "M" {
				mem = mem * 1000
			}
			pid_map[pid] = process_information{pid, ppid, cpu, mem}
			if _, found := children_map[ppid]; found {
				children_map[ppid] = children{append(children_map[ppid].pids, pid)}
			} else {
				var slice []string
				slice = append(slice, pid)
				children_map[ppid] = children{slice}
			}
		}
	}
	return pid_map, children_map, nil
}
