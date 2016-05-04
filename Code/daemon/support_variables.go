//filename: support_variables.go
//information: created on 8th of September 2015 by Andreas Kittilsland
//Contains global variables and structures

package main

import (
	"io"
)
var test_applications = false
var target_size = 20        //The max allowed size of the network
const emulate_network = false //if true, all communication between daemons go by the network emulator

var SimpleTxt_port = "8500" //Port to access the interface of SimpleTxt, our text editor

var coord_ver_vector_mu = ExpandedLock{} //Mutex for the version vector at the coordinator (lock.go)
var coord_version_vector map[int]int     //Version vector at the coordinator
var ver_vector_mu = ExpandedLock{}       //Mutex for the version vector (lock.go)
var version_vector map[int]int           //Version vector

//Locked if the local and coordinator version vector was found to be different on connect.
//Disallows local instigated changes, but incoming changes from the coordinator is still allowed
var initialize = ExpandedLock{}

var refreshing = ExpandedLock{} //Lock for when the system is refreshing all the daemons. Useful if new daemons connect during refreshing

var pid_string = "0"
var chg_trk_pid_string = "0"

var ChangeDetector_InPipe io.WriteCloser

var auto_coord bool //true if we are to run automatic coordinator selection
