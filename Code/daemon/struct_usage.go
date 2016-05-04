//filename: struct_usage.go
//information: created on 8th of January 2016 by Andreas Kittilsland
package main

import (
	"time"
)

var usage_m = ExpandedLock{}
var nr_of_runs = 0
var nr_of_changes = 0
var nr_local_objs = 0 //Number of local objects

var reset_interval = time.Minute * 5

type Queue struct {
	list  []time.Time
	types []string
}

var metrics_times = Queue{}

func (q *Queue) Push(t string) {
	usage_m.Lock()
	if t == "run" {
		nr_of_runs++
	} else {
		nr_of_changes++
	}
	q.list = append(q.list, time.Now())
	q.types = append(q.types, t)
	usage_m.Unlock()
}

func (q *Queue) Pop() {
	usage_m.Lock()
	q.list = q.list[1:]
	q.types = q.types[1:]
	usage_m.Unlock()
}

//Must be sure to lock when using peeks, as they have NO LOCKS IN THEM....
func (q *Queue) TimePeek() time.Time {
	if len(q.list) > 0 {
		ret_val := q.list[0]
		return ret_val
	} else {
		return time.Now()
	}
}

func (q *Queue) TypePeek() string {
	if len(q.types) > 0 {
		ret_val := q.types[0]
		return ret_val
	} else {
		return "N/A"
	}
}

func ResetMetrics() {
	usage_m.Lock()
	for time.Since(metrics_times.TimePeek()) > reset_interval {
		if metrics_times.TypePeek() == "run" {
			nr_of_runs--
		} else {
			nr_of_changes--
		}
		metrics_times.Pop()

	}
	usage_m.Unlock()
}
