//filename: func_gui.go
//information: created on 22th of September 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
)

var http_server http.Server
var http_gui_port = "0"

//Initilize and listen for browsers requests for the graphics
func run_gui() {
	http.HandleFunc("/refresh_objs", gui_RefreshObjs)
	http.HandleFunc("/refresh_apps", gui_RefreshApps)
	http.HandleFunc("/open_file", gui_OpenFile)
	http.HandleFunc("/add_file", gui_AddFile)
	http.HandleFunc("/delete_file", gui_DeleteFile)
	http.HandleFunc("/change_name", gui_ChangeName)
	http.HandleFunc("/applications", gui_ShowApplication)
	http.HandleFunc("/", gui_ShowMain)
	http.ListenAndServe(":"+http_gui_port, nil)
}

func gui_RefreshObjs(w http.ResponseWriter, r *http.Request) {
	//rc_* functions are in remote_commands.go
	rc_RequestRefresh("objs")
	gui_ShowMain(w, r)
}

func gui_RefreshApps(w http.ResponseWriter, r *http.Request) {
	//rc_* functions are in remote_commands.go
	rc_RequestRefresh("apps")
	gui_ShowApplication(w, r)
}

func gui_OpenFile(w http.ResponseWriter, r *http.Request) {
	//Parse information passed
	r.ParseForm()
	id := r.Form["id"][0]

	conn, err := net.Dial("tcp", ip_self+":"+port)
	ErrorCheck(err, true)
	defer conn.Close()

	WriteLines(conn, "client open file", id)

	conn_reader := bufio.NewReader(conn)
	result := ReadLine(conn_reader)
	if result == "found" {
		url := ReadLine(conn_reader)
		color_tag := "<font color=\"red\">"
		fmt.Fprintf(w, "<html>")
		fmt.Fprintf(w, "<head>")
		fmt.Fprintf(w, "<meta http-equiv=\"refresh\" content=\"1; url=http://"+url+"\">")
		fmt.Fprintf(w, "<title>Remote File Interface</title>")
		fmt.Fprintf(w, "</head>")
		fmt.Fprintf(w, "<body bgcolor=\"black\">")
		fmt.Fprintf(w, color_tag+"<center><h2>Opening remote file...</h2></center>")
		fmt.Fprintf(w, "</font></body></html>")
	} else if result == "ok" {
		url := ReadLine(conn_reader)
		color_tag := "<font color=\"red\">"
		fmt.Fprintf(w, "<html>")
		fmt.Fprintf(w, "<head>")
		fmt.Fprintf(w, "<title>Remote File Interface</title>")
		fmt.Fprintf(w, "<script type=\"text/javascript\">setTimeout(function ( ){  self.close();}, 500 );</script>")
		fmt.Fprintf(w, "</head>")
		fmt.Fprintf(w, "<body bgcolor=\"black\">")
		fmt.Fprintf(w, color_tag+"<center><iframe src=\"http://"+url+"\"></iframe></center>")
		fmt.Fprintf(w, "</font></body></html>")
	} else {
		color_tag := "<font color=\"red\">"
		fmt.Fprintf(w, "<html>")
		fmt.Fprintf(w, "<head>")
		fmt.Fprintf(w, "<title>Error</title>")
		fmt.Fprintf(w, "</head>")
		fmt.Fprintf(w, "<body style=\"width: 100%; height: 100%\" bgcolor=\"black\"><center>")
		fmt.Fprintf(w, color_tag+"<h2>Error: File was either not found, or was already open</h2>")
		fmt.Fprintf(w, "<button type=\"button\" onclick=\"window.open('', '_self', ''); window.close();\">Close</button>")
		fmt.Fprintf(w, "</center></font></body></html>")
	}
}

func gui_AddFile(w http.ResponseWriter, r *http.Request) {
	//Parse information passed
	r.ParseForm()
	fname := r.Form["file_name"][0]

	wd := SetWD(daemon_dir)

	f, err := os.Create("../files/daemon_" + port + "/" + fname)
	ErrorCheck(err, true)
	f.Close()

	ResetWD(wd)

	ToChangeTracker("add\n" + "daemon_" + port + "/" + fname + "\n") //func_main.go

	color_tag := "<font color=\"red\">"
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<meta http-equiv=\"refresh\" content=\"0; url=http://"+ip_self+":"+http_gui_port+"/\">")
	fmt.Fprintf(w, "<title>Refreshing Interface</title>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body bgcolor=\"black\">")
	fmt.Fprintf(w, color_tag+"<center><h2>Refreshining interface...</h2></center>")
	fmt.Fprintf(w, "</font></body></html>")
}

func gui_DeleteFile(w http.ResponseWriter, r *http.Request) {
	//Parse information passed
	r.ParseForm()
	id := r.Form["id"][0]
	id_int, err := strconv.Atoi(id)
	ErrorCheck(err, true)

	err = SendDeletionRequest(id_int)
	ErrorCheck(err, true)

	color_tag := "<font color=\"red\">"
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<meta http-equiv=\"refresh\" content=\"0; url=http://"+ip_self+":"+http_gui_port+"/\">")
	fmt.Fprintf(w, "<title>Refreshing Interface</title>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body bgcolor=\"black\">")
	fmt.Fprintf(w, color_tag+"<center><h2>Refreshining interface...</h2></center>")
	fmt.Fprintf(w, "</font></body></html>")
}

func gui_ChangeName(w http.ResponseWriter, r *http.Request) {
	//Parse information passed
	r.ParseForm()
	fname := r.Form["fname"][0]
	pfname := r.Form["pfname"][0]

	wd := SetWD(daemon_dir)

	os.Rename("../files/daemon_"+port+"/"+pfname, "../files/daemon_"+port+"/"+fname)

	ResetWD(wd)

	ToChangeTracker("name_change\n" + "daemon_" + port + "/" + pfname + ":daemon_" + port + "/" + fname + "\n")

	color_tag := "<font color=\"red\">"
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<meta http-equiv=\"refresh\" content=\"0; url=http://"+ip_self+":"+http_gui_port+"/\">")
	fmt.Fprintf(w, "<title>Refreshing Interface</title>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body bgcolor=\"black\">")
	fmt.Fprintf(w, color_tag+"<center><h2>Refreshining interface...</h2></center>")
	fmt.Fprintf(w, "</font></body></html>")
}

//Shows the graphics on the tile through HTML
func gui_ShowApplication(w http.ResponseWriter, r *http.Request) {
	font_colour := "red"
	color_tag := "<font color=\"" + font_colour + "\">"
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<title>Daemon Interface</title>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body bgcolor=\"black\">")
	fmt.Fprintf(w, "<center>"+color_tag)
	fmt.Fprintf(w, "<p><b> Running on "+ip_self+":"+port+", with PID: "+pid_string+" </b></p>")
	if is_coord {
		fmt.Fprintf(w, "<p>I am the coordinator daemon</p>")
	} else {
		fmt.Fprintf(w, "<p>I am a profane daemon</p>")
	}
	fmt.Fprintf(w, "<b> Applications: </b>")
	applications_mu.ReadOnly()
	fmt.Fprintf(w, "<form action=\"/refresh_apps\"><input type=\"submit\" value=\"Refresh\"></form>")
	if len(applications) > 0 {
		if len(applications) == 1 {
			fmt.Fprintf(w, "I have 1 application</br></br>")
		} else {
			fmt.Fprintf(w, "I have "+strconv.Itoa(len(applications))+" applications</br></br>")
		}
		fmt.Fprintf(w, "<form action=\"/\"><input type=\"submit\" value=\"Go to to objects list\"></form>")
		for _, app := range applications {
			fmt.Fprintf(w, app.app_name+"</br>")
		}
	} else {
		fmt.Fprintf(w, "I have no applications")
	}
	applications_mu.Editable()

	fmt.Fprintf(w, "</font></center>")
	fmt.Fprintf(w, "</html>")
}

//Shows the graphics on the tile through HTML
func gui_ShowMain(w http.ResponseWriter, r *http.Request) {
	font_colour := "red"
	color_tag := "<font color=\"" + font_colour + "\">"
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<title>Daemon Interface</title>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body bgcolor=\"black\">")
	fmt.Fprintf(w, "<center>"+color_tag)
	fmt.Fprintf(w, "<p><b> Running on "+ip_self+":"+port+", with PID: "+pid_string+" </b></p>")
	if is_coord {
		fmt.Fprintf(w, "<p>I am the coordinator daemon</p>")
	} else {
		fmt.Fprintf(w, "<p>I am a profane daemon</p>")
	}
	var applications_string string
	if applications_mu.IsLocked() || local_apps_mu.IsLocked() {
		applications_string = "I have ca. " + strconv.Itoa(len(applications)) + "  applications"
	} else {
		applications_mu.ReadOnly()
		if len(applications) > 1 {
			applications_string = "I have " + strconv.Itoa(len(applications)) + "  applications"
		} else if len(applications) == 1 {
			applications_string = "I have 1 application"
		} else {
			applications_string = "I have no applications"
		}
		applications_mu.Editable()
	}
	obj_mu.ReadOnly()
	if len(objects) > 0 {
		if len(objects) == 1 {
			fmt.Fprintf(w, "I have 1 object and "+applications_string+"</br></br>")
		} else {
			fmt.Fprintf(w, "I have "+strconv.Itoa(len(objects))+" objects and "+applications_string+"</br></br>")
		}
	} else {
		fmt.Fprintf(w, "I have no objects and "+applications_string+"</br></br>")
	}
	fmt.Fprintf(w, "<form action=\"/add_file\">Create a new file with title: <input type=\"text\" name=\"file_name\">"+
		"<input type=\"submit\" value=\"Create\"></form>")
	fmt.Fprintf(w, "<b> Objects: </b>")
	fmt.Fprintf(w, "<form action=\"/refresh_objs\"><input type=\"submit\" value=\"Refresh\"></form>")
	if len(objects) > 0 {
		fmt.Fprintf(w, "<form action=\"/applications\"><input type=\"submit\" value=\"Go to applications list\"></form>")
		fmt.Fprintf(w, "<table id=\"TBL\" style=\"width:100%\" border=\"1\">")
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, "<th>"+color_tag+"Open File</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Delete File</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Type</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"ID</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Version</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Name</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Modification Date</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Size</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Codec</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Resolution</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Format</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"FPS</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Sound Quality</font></th>")
		fmt.Fprintf(w, "<th>"+color_tag+"Sample Rate</font></th>")
		fmt.Fprintf(w, "</tr>")

		columns := map[string]string{
			"name":       "N/A",
			"mod_date":   "N/A",
			"size":       "N/A",
			"codec":      "N/A",
			"resolution": "N/A",
			"format":     "N/A",
			"fps":        "N/A",
			"quality":    "N/A",
			"sample":     "N/A",
		}
		f_obj := 0
		for i := 0; f_obj < len(objects); i++ {
			if c_object, exists := objects[i]; exists {
				f_obj++
				fmt.Fprintf(w, "<tr>")
				fmt.Fprintf(w, "<td><center><form action=\"/open_file\" target=\"_blank\"><input hidden type=\"text\""+
					"name=\"id\" value=\""+strconv.Itoa(i)+"\"><input type=\"submit\" value=\"Open\"></form></center></td>")
				fmt.Fprintf(w, "<td><center><form action=\"/delete_file\"><input hidden type=\"text\""+
					"name=\"id\" value=\""+strconv.Itoa(i)+"\"><input type=\"submit\" value=\"Delete\"></form></center></td>")
				fmt.Fprintf(w, "<td><center>"+color_tag+c_object.obj_type+"</font></center></td>")
				fmt.Fprintf(w, "<td><center>"+color_tag+strconv.Itoa(i)+"</font></center></td>")
				fmt.Fprintf(w, "<td><center>"+color_tag+strconv.Itoa(c_object.version)+"</font></center></td>")
				for i := 0; i < len(columns); i++ {
					var index string
					if i == 0 {
						index = "name"
					} else if i == 1 {
						index = "mod_date"
					} else if i == 2 {
						index = "size"
					} else if i == 3 {
						index = "codec"
					} else if i == 4 {
						index = "resolution"
					} else if i == 5 {
						index = "format"
					} else if i == 6 {
						index = "fps"
					} else if i == 7 {
						index = "quality"
					} else if i == 8 {
						index = "sample"
					}
					if val, found := c_object.attributes[index]; found {
						if index == "name" {
							fmt.Fprintf(w, "<td><center><form action=\"/change_name\"><input hidden type=\"text\" name=\"pfname\" value=\""+val+"\"></input><input type=\"text\" name=\"fname\" value=\""+val+"\"></input><input type=\"submit\" value=\"Save\"></form></center></td>")
						} else {
							fmt.Fprintf(w, "<td><center>"+color_tag+val+"</font></center></td>")
						}
					} else {
						fmt.Fprintf(w, "<td><center>"+color_tag+columns[index]+"</font></center></td>")
					}
				}
				fmt.Fprintf(w, "</tr>")
			}
		}
		fmt.Fprintf(w, "</table>")
	}
	obj_mu.Editable()

	fmt.Fprintf(w, "</font></center>")
	fmt.Fprintf(w, "</html>")
}
