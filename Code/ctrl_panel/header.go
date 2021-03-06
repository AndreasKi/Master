//filename: client.go
//information: created on 22th of September 2015 by Andreas Kittilsland
package main

import (
	"net/http"
	"sync"
	"time"
)

type daemon struct { //Structure for the details of the list of daemons
	ip, port, pid string
}

var daemons_mu = sync.Mutex{} //Mutex for the list of daemons in our network
var daemons []daemon          //List of daemons in our network

var ip = "localhost"
var port = ":8580"

var pid_string = "N/A"

var http_server http.Server
var http_gui_port = ":8570"

var timing time.Time

var gui_notification []string

var target_size = 20
