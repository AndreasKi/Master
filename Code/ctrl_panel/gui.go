//filename: gui.go
//information: created on 22th of September 2015 by Andreas Kittilsland
package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

//Initilize and listen for browsers requests for the graphics
func run_gui() {
	http.HandleFunc("/", show_main_gui)
	http.HandleFunc("/add_daemon", gui_add_daemon)
	http.HandleFunc("/change_settings", gui_set_vars)
	http_listener, err := net.Listen("tcp", http_gui_port)
	ErrorCheck(err, true)
	for true {
		err := http_server.Serve(http_listener)
		ErrorCheck(err, true)
	}
}

//Shows the graphics on the client
func show_main_gui(w http.ResponseWriter, r *http.Request) {
	check_connectivity()
	var mems []string
	var cpus []string

	font_colour := "blue"
	color_tag := "<font color=\"" + font_colour + "\">"

	fmt.Fprintf(w, "<!DOCTYPE html>")
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<title>Control Panel</title>")
	//Script allowing for automatic resizing of the iframes to fit the page
	fmt.Fprintf(w,
		"<script language=\"javascript\" type=\"text/javascript\">"+
			"function resizeIframe(obj) {"+
			"obj.style.height = obj.contentWindow.document.body.scrollHeight + 'px';"+
			"}"+
			"</script>")
	fmt.Fprintf(w, "<style type=\"text/css\"> html, body { height: 100%; width: 100%; }</style>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body bgcolor=\"black\">")
	fmt.Fprintf(w, "<center><font color=\""+font_colour+"\">")
	fmt.Fprintf(w, "<h1> Control Panel </h1>")
	fmt.Fprintf(w, "<h3> Running on "+ip+port+", with PID: "+pid_string+" </h3>")
	fmt.Fprintf(w, "<h3> CPU usage: "+strconv.FormatFloat(get_cpu(pid_string), 'f', 1, 64)+" &#37; </h3>")
	fmt.Fprintf(w, "<h3> Memory usage: "+strconv.FormatFloat(get_memory(pid_string), 'f', 1, 64)+" &#37; </h3>")

	for _, n := range gui_notification {
		fmt.Fprintf(w, "<p>"+n+"</p>")
	}
	fmt.Fprintf(w, "</br><h3>Change Settings</h3>")
	fmt.Fprintf(w, "<form action=\"/change_settings\"><b>When electing coordinator, take into account values as old as:</b></br>"+
		"Minutes: <input type=\"text\" name=\"interval_min\" value=\"5\">\t"+
		"Seconds: <input type=\"text\" name=\"interval_sec\" value=\"0\"></br>"+
		"</br><b>How to weigh the different variables when electing coordinator:</b></br>"+
		"Number of Files: <input type=\"text\" name=\"files\" value=\"1\"></br>"+
		"Number of Application Runs: <input type=\"text\" name=\"runs\" value=\"10\"></br>"+
		"Number of Changes and Updates: <input type=\"text\" name=\"changes\" value=\"5\"></br>"+
		"</br>Where to set the battery threshold: <input type=\"text\" name=\"battery\" value=\"20\"> &#37;</br>"+
		"</br>How often to evaluate the coordinator: <input type=\"text\" name=\"eval\" value=\"5\"> seconds</br>"+
		"<input type=\"submit\" value=\"Apply\"></form></br>")

	fmt.Fprintf(w, "</br><h3>Daemons</h3>")
	daemons_mu.Lock()
	total_cpu := 0.0
	total_mem := 0.0
	for _, cur_d := range daemons {
		if cur_d.pid != "0" {
			cpu := get_memory(cur_d.pid)
			mem := get_cpu(cur_d.pid)

			total_cpu = total_cpu + cpu
			total_mem = total_mem + mem

			cpus = append(cpus, strconv.FormatFloat(cpu, 'f', 1, 64))
			mems = append(mems, strconv.FormatFloat(mem, 'f', 1, 64))
		} else {
			cpus = append(cpus, "N/A")
			mems = append(mems, "N/A (Refresh)")
		}
	}
	fmt.Fprintf(w, "Currently running: "+strconv.Itoa(len(daemons))+"</br>")
	fmt.Fprintf(w, "Approx. Total CPU Usage: "+strconv.FormatFloat(total_cpu, 'f', 1, 64)+" &#37;</br>")
	fmt.Fprintf(w, "Approx. Total Memory Usage: "+strconv.FormatFloat(total_mem, 'f', 1, 64)+" &#37;</br>")
	fmt.Fprintf(w, "<form action=\"/add_daemon\"><input type=\"submit\" value=\"Add Daemon\"></form>")
	fmt.Fprintf(w, "<form action=\"/\"><input type=\"submit\" value=\"Check Connectivity to Daemons\"></form>")

	//Show the pages of all the daemons
	fmt.Fprintf(w, "<table id=\"DMNTBL\" style=\"table-layout: fixed;\" border=\"0\">")
	row_len := 1
	for index, cur_daemon := range daemons {
		if row_len == 1 {
			fmt.Fprintf(w, "<tr>")
		}
		cur_daemon_port_int, err := strconv.Atoi(cur_daemon.port)
		ErrorCheck(err, true)
		link_to_daemon := "http://" + cur_daemon.ip + ":" + strconv.Itoa((cur_daemon_port_int-8590)+9000)
		fmt.Fprintf(w, "<td><b>"+color_tag+"Daemon "+strconv.Itoa(index+1)+" <a href=\""+link_to_daemon+"\" target=\"_blank\">View Objects</a> "+
			"<a href=\""+link_to_daemon+"/applications\" target=\"_blank\">View Applications</a></br>")
		fmt.Fprintf(w, "CPU usage: "+cpus[index]+" &#37; Memory usage: "+mems[index]+" &#37; </font></b></br>")
		fmt.Fprintf(w, "<iframe src=\""+link_to_daemon+"\" frameborder=\"1\" scrolling=\"no\" onload='javascript:resizeIframe(this);'></iframe>")
		if row_len == 3 {
			fmt.Fprintf(w, "</tr>")
			row_len = 1
		} else {
			row_len++
		}
	}
	fmt.Fprintf(w, "</table>")
	daemons_mu.Unlock()

	fmt.Fprintf(w, "</font></center>")
	fmt.Fprintf(w, "</html>")
}
