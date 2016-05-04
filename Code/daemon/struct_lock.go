//filename: struct_lock.go
//information: created on 23th of September 2015 by Andreas Kittilsland

package main

import (
	"sync"
)

//Main lock structure
type ExpandedLock struct {
	totally_locked bool         //status of the lock
	read_only      int          //Number of read only locks triggered.
	lock           sync.RWMutex //The lock
}

//Locks object completely. Even reading disallowed
func (l *ExpandedLock) Lock() {
	l.lock.Lock()
	l.totally_locked = true
}

//Unlocks a locked object
func (l *ExpandedLock) Unlock() {
	l.totally_locked = false
	l.lock.Unlock()
}

//Locks object from being edited, reads allowed
func (l *ExpandedLock) ReadOnly() {
	l.lock.RLock()
	l.read_only++
}

//Removes edit-hindering lock
func (l *ExpandedLock) Editable() {
	l.read_only--
	l.lock.RUnlock()
}

//Returns true if the lock is locked, else false
func (l ExpandedLock) IsLocked() bool {
	return l.totally_locked
}

//Returns true if the lock is read only, else false
func (l ExpandedLock) IsReadOnly() bool {
	return l.read_only > 0
}
