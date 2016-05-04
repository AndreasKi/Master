//filename: struct_update-queue.go
//information: created on 10th of March 2016 by Andreas Kittilsland

package main

type UpdateQueue struct {
	msgs []string
}

//Local non-pushed oupdates
var updates_m = ExpandedLock{}
var WaitingUpdates = UpdateQueue{}

func (q *UpdateQueue) Push(msg string) {
	updates_m.Lock()
	q.msgs = append(q.msgs, msg)
	updates_m.Unlock()
}

func (q *UpdateQueue) Pop() string {
	updates_m.Lock()
	ret_val := q.msgs[0]
	q.msgs = q.msgs[1:]
	updates_m.Unlock()

	return ret_val
}

func (q *UpdateQueue) Peek() string {
	updates_m.ReadOnly()
	if len(q.msgs) > 0 {
		ret_val := q.msgs[0]
		updates_m.Editable()
		return ret_val
	} else {
		updates_m.Editable()
		return "nil"
	}
}
