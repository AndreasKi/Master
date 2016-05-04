//filename: gui.go
//information: created on 22th of September 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"net/http"
	"net"
	"fmt"
	"os/exec"
	"errors"
	"strings"
)

func gui_add_daemon(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("../daemon/daemon")

	cmdReader, err := cmd.StdoutPipe()
	ErrorCheck(err, true)

	scanner := bufio.NewScanner(cmdReader)

	err = cmd.Start()
	ErrorCheck(err, true)

	ip_d := "N/A"
	port_d := "N/A"
	d_pid := "N/A"
	if scanner.Scan() {
		line := scanner.Text()
		daemon_info := strings.Split(line, ":")
		ip_d = daemon_info[0]
		port_d = daemon_info[1]
		d_pid = daemon_info[2]
	}
	new_daemon := daemon{ip_d, port_d, d_pid}

	daemons_mu.Lock()
	daemons = append(daemons, new_daemon)
	daemons_mu.Unlock()

	go read_daemon_output(scanner, ip_d, port_d)

	AddNotification("A new daemon was successfully added")

	show_main_gui(w, r)
}

func gui_set_vars(w http.ResponseWriter, r *http.Request) {
	interval_min := r.FormValue("interval_min")
	interval_sec := r.FormValue("interval_sec")
	files := r.FormValue("files")
	runs := r.FormValue("runs")
	changes := r.FormValue("changes")
	battery := r.FormValue("battery")
	eval := r.FormValue("eval")

	conn, err := net.Dial("tcp", "127.0.0.1:"+daemons[0].port)
	ErrorCheck(err, true)

	defer conn.Close()
	conn_reader := bufio.NewReader(conn)
	fmt.Fprintf(conn, "set vars\n"+interval_min+"\n"+interval_sec+"\n"+files+"\n"+runs + "\n" + changes+"\n"+battery+"\n"+eval+"\n")
	resp, err := conn_reader.ReadString('\n')
	ErrorCheck(err, true)

	resp = resp[:len(resp)-1]
	if resp != "ack" {
		ErrorCheck(errors.New("Failed to comprehend response from daemon"), true)
	}

	AddNotification("Settings changed")

	show_main_gui(w, r)
}

func AddNotification(notification string) {
	if len(gui_notification) >= 3 {
		gui_notification = gui_notification[1:]
	}
	gui_notification = append(gui_notification, notification)
	return
}
