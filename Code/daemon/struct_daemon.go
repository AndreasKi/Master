//filename: support_daemon.go
//information: created on 8th of September 2015 by Andreas Kittilsland
//Contains global variables and structures

package main

type daemon struct { //Structure for the details of the list of daemons
	ip, port string
}

var daemons_mu = ExpandedLock{} //Mutex for the list of daemons in our network
var daemons []daemon            //List of daemons in our network
