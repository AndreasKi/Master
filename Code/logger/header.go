package main

import (
	"sync"
	"time"
)

type daemon struct {
	ip, port, pid string
	dead          bool      //true if daemon is not running anymore
	cpu_history   []float64 //holds histories of cpu and memory usage
	mem_history   []int
	start_time    int //How many seconds after start did we find this daemon
}

var daemons_mu = sync.Mutex{} //Mutex for the list of daemons
var daemons []daemon          //List of daemons that ran during the test

var target_size = 20

var start time.Time

var log_time = -1 //How many secs to run logging for

var stop = false
var finished = sync.WaitGroup{}

var interval = 3
