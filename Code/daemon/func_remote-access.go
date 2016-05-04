//filename: func_remote-access.go
//information: created on 9th of October 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var registered map[int]bool //List of which objects we have exposed the HTML interface for
var logText map[int]string  //Message displayed on the HTML GUI
var loaded_scripts = ""     //Javascripts read from javascript.js used in the header of the HTML GUI
var prev_txt map[int]string //The text content of the object at previous check. Used to keep track of changes made by the user.
var compatible map[int]bool //Whether or not the object is the correct object type
var is_running map[int]bool //Instantiated in init.go
var page map[int]int
var page_size = 750 //Number of chars per page, rounded up or down to a whole word

//Initilize and listen for browsers requests for the graphics
func ExposeFileInterface(id int, port_to_use string) {
	if len(registered) == 0 {
		registered = make(map[int]bool)
		logText = make(map[int]string)
		prev_txt = make(map[int]string)
		compatible = make(map[int]bool)
		page = make(map[int]int)

		//Read the javascripts we need for the HTML interface
		bytes, err := ioutil.ReadFile("javascript.js")
		ErrorCheck(err, true)
		loaded_scripts = string(bytes)
	}

	id_string := strconv.Itoa(id)
	//Check that we didnt already set it up.
	if !registered[id] { //First time opening this object, register the handler for this object.
		if objects[id].extension != "stxt" {
			http.HandleFunc("/"+id_string, ShowErr)
			compatible[id] = false
		} else {
			//Expose the HTML interface
			http.HandleFunc("/"+id_string, ShowFile)
			logText[id] = "No errors"
			http.HandleFunc("/save_file_"+id_string, SaveFile)
			http.HandleFunc("/up_"+id_string, Up)
			http.HandleFunc("/down_"+id_string, Down)
			http.HandleFunc("/make_bold_"+id_string, MakeBold)
			http.HandleFunc("/make_italic_"+id_string, MakeItalic)
			http.HandleFunc("/make_underlined_"+id_string, MakeUnderlined)
			http.HandleFunc("/editor_"+id_string, EditMonitor)
			http.HandleFunc("/exit_"+id_string, Exit)
			compatible[id] = true

			page[id] = 0
			RunApplication(id)
		}
		registered[id] = true
	}

	//Run the application if the file was opened and closed previously
	if registered[id] && !is_running[id] && compatible[id] {
		RunApplication(id)
	}

	//Tell the application to open the file
	if is_running[id] && registered[id] {
		locks[id].ReadOnly()
		c_wd := SetWD(daemon_dir)

		fpath, err := filepath.Abs(std_dir + objects[id].file_path)
		ErrorCheck(err, true)

		ResetWD(c_wd)

		message := "open\n" + fpath + "/" + objects[id].attributes["name"] + "\n"

		if err = DialAndSend(ip_self, GetAppPort(id), message); ErrorCheck(err) {
			logText[id] = "Failed to open file - " + string(err.Error())
		} else {
			logText[id] = "No errors"
		}
		locks[id].Editable()
	}

	go http.ListenAndServe(":"+port_to_use, nil)
}

func SaveFile(w http.ResponseWriter, r *http.Request) {
	id := ReadIDFromURL(r)

	//Overwrite content of old file with new content
	message := "save\n"
	if err := DialAndSend(ip_self, GetAppPort(id), message); ErrorCheck(err) {
		logText[id] = "Failed to commit changes to file - " + string(err.Error())
	} else {
		logText[id] = "File was succesfully saved!"
	}

	//Show the interface again
	ShowFile(w, r)
}

//Move the view up
func Up(w http.ResponseWriter, r *http.Request) {
	id := ReadIDFromURL(r)
	page[id] = page[id] - 1
	ShowFile(w, r)
}

//Move the view down
func Down(w http.ResponseWriter, r *http.Request) {
	id := ReadIDFromURL(r)
	page[id] = page[id] + 1
	ShowFile(w, r)
}

func TextOperation(w http.ResponseWriter, r *http.Request, op string) {
	id := ReadIDFromURL(r)

	r.ParseForm()
	//full := r.Form["text_content"][0]
	selection := r.Form["selected_text"][0]
	pos := r.Form["pos"][0]

	startPos, err := strconv.Atoi(pos)
	ErrorCheck(err, true)
	endPos := startPos + len(selection)

	//Send positions
	message := op + "\n" + strconv.Itoa(startPos) + "\n" + strconv.Itoa(endPos) + "\n"
	if err := DialAndSend(ip_self, GetAppPort(id), message); ErrorCheck(err) {
		logText[id] = "Failed to commit changes to file - " + string(err.Error())
	} else {
		logText[id] = "Text was made " + op
	}

	//Show the interface again
	ShowFile(w, r)
}

func MakeBold(w http.ResponseWriter, r *http.Request) {
	TextOperation(w, r, "bold")
}

func MakeItalic(w http.ResponseWriter, r *http.Request) {
	TextOperation(w, r, "italic")
}

func MakeUnderlined(w http.ResponseWriter, r *http.Request) {
	TextOperation(w, r, "underlined")
}

//Looks for changes made in the text by the user
func EditMonitor(w http.ResponseWriter, r *http.Request) {
	id := ReadIDFromURL(r)
	diff_text := r.FormValue("text_content")
	diff_point := r.FormValue("diff_point")
	diff_end := r.FormValue("equal_point")

	diff_text = strings.Replace(diff_text, "<--endl-->", ";", -1)
	diff_text = strings.Replace(diff_text, "\n", "<--newline-->", -1)

	if err := DialAndSend(ip_self, GetAppPort(id), "update\n"+diff_point+"\n"+diff_end+"\n"+diff_text+"\n<--done-->\n"); ErrorCheck(err) {
		logText[id] = "Synchronization with the application failed!"
	}
}

//Shut down the server, and redirect the client back to the object list
func ShowErr(w http.ResponseWriter, r *http.Request) {
	color_tag := "<font color=\"yellow\">"
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<title>Error</title>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body style=\"width: 100%; height: 100%\" bgcolor=\"black\"><center>")
	fmt.Fprintf(w, color_tag+"<h2>Error: Remote access has not been implemented for this file type.</h2>")
	fmt.Fprintf(w, "<button type=\"button\" onclick=\"window.open('', '_self', ''); window.close();\">Close</button>")
	fmt.Fprintf(w, "</center></font></body></html>")
}

//Shut down
func Exit(w http.ResponseWriter, r *http.Request) {
	id := ReadIDFromURL(r)

	err := DialAndSend(ip_self, GetAppPort(id), "exit\n")
	if !ErrorCheck(err) {
		is_running[id] = false
		w.Write([]byte("ack"))
	}
	//RunWatchdogScan(true)
}

//Shows the graphics on the tile through HTML
func ShowFile(w http.ResponseWriter, r *http.Request) {
	id := ReadIDFromURL(r)
	locks[id].ReadOnly()

	//Open and read the content of the requested file
	//dat, err := ioutil.ReadFile(std_dir + objects[id].file_path + objects[id].attributes["name"])
	//ErrorCheck(err, true)

	start_pos := page_size * page[id]
	editor_conn, err := Dial(ip_self, GetAppPort(id))
	ErrorCheck(err, true)
	WriteLines(editor_conn, "get", strconv.Itoa(start_pos), strconv.Itoa(page_size))
	editor_reader := bufio.NewReader(editor_conn)
	txt := ReadDocumentFragment(editor_reader)

	is_running[id] = true

	//Insert object ID into the javascripts where necessary
	scripts := strings.Replace(loaded_scripts, "<!--ObjID--!>", strconv.Itoa(id), -1)
	//txt := string(dat)

	//Show the interface
	color_tag := "<font color=\"yellow\">"
	name := objects[id].attributes["name"]
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<title>Text Editor - "+name+"</title>")
	fmt.Fprintf(w, "<style type=\"text/css\"> div { background-color:white;} </style>")
	fmt.Fprintf(w, "<script type=\"text/javascript\">"+scripts+"</script>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body bgcolor=\"black\">")
	fmt.Fprintf(w, color_tag+"<center><h3 id=\"julenisse\">SimpleTxt Remote Interface - "+name+"</h3>")

	fmt.Fprintf(w, "<button type=\"button\" onclick=\"exitEditor();\">Close</button>"+
		"<button type=\"button\" onclick=\"location.href='save_file_"+strconv.Itoa(id)+"';\" id=\"save\">Save</button></br>&nbsp;"+
		"<button type=\"button\" onclick=\"makeBold();\" id=\"bold\"><b>Bold</b></button>"+
		"<button type=\"button\" onclick=\"makeItalic();\" id=\"italic\"><i>Italic</i></button>"+
		"<button type=\"button\" onclick=\"makeUnderlined();\" id=\"underlined\"><u>Underlined</u></button>"+
		"<font color=\"black\"><div class=\"textarea\" align=\"left\" id=\"textBox\" contenteditable=\"true\" style=\"width:600px;"+
		"\">"+txt+"</div></font>"+
		"<button style=\"background-color:lightblue\" type=\"button\" onclick=\"location.href='up_"+strconv.Itoa(id)+"';\" id=\"up\">Previous Page</button>"+
		"<button style=\"background-color:lightgreen\" type=\"button\" onclick=\"location.href='down_"+strconv.Itoa(id)+"';\" id=\"up\">Next Page</button></br>&nbsp;")
	fmt.Fprintf(w, "<p>"+logText[id]+"</p></center></font>")
	fmt.Fprintf(w, "<div hidden id =\"obj_id\">"+strconv.Itoa(id)+"</div>")
	fmt.Fprintf(w, "<div hidden id =\"prev_text\" class=\"textarea\"></div>")

	fmt.Fprintf(w, "<form action=\"make_bold_"+strconv.Itoa(id)+"\" id=\"makebold\">"+
		"<textarea hidden id=\"selectedtext_bold\" name=\"selected_text\"></textarea>"+
		"<textarea hidden id=\"pos_bold\" name=\"pos\"></textarea>"+
		"<textarea hidden id=\"fulltext_bold\" name=\"text_content\"></textarea></form>")
	fmt.Fprintf(w, "<form action=\"make_italic_"+strconv.Itoa(id)+"\" id=\"makeitalic\">"+
		"<textarea hidden id=\"selectedtext_italic\" name=\"selected_text\"></textarea>"+
		"<textarea hidden id=\"pos_italic\" name=\"pos\"></textarea>"+
		"<textarea hidden id=\"fulltext_italic\" name=\"text_content\"></textarea></form>")
	fmt.Fprintf(w, "<form action=\"make_underlined_"+strconv.Itoa(id)+"\" id=\"makeunderlined\">"+
		"<textarea hidden id=\"selectedtext_underlined\" name=\"selected_text\"></textarea>"+
		"<textarea hidden id=\"pos_underlined\" name=\"pos\"></textarea>"+
		"<textarea hidden id=\"fulltext_underlined\" name=\"text_content\"></textarea></form>")

	fmt.Fprintf(w, "</body>")
	fmt.Fprintf(w, "</html>")
	locks[id].Editable()
}

//Parses the URL from the request to get the object ID
func ReadIDFromURL(r *http.Request) int {
	//Remove any information on forms etc.
	url_c := strings.Split(r.URL.String(), "?")

	//Remove the ip and port
	url := strings.Split(url_c[0], "/")

	//Remove the function name/subdir
	subdir := strings.Split(url[len(url)-1], "_")
	id_string := subdir[len(subdir)-1]

	//Convert the id to integer from string
	id_int, err := strconv.Atoi(id_string)
	ErrorCheck(err, true, true, true)

	return id_int
}

func RunApplication(id int) {
	//Run the application
	if !is_running[id] {
		c_wd := SetWD(daemon_dir)
		cmd := exec.Command("../SimpleTxt/SimpleTxt", strconv.Itoa(id))
		wd := daemon_dir
		wd = wd[:len(wd)-6]
		cmd.Dir = wd + "SimpleTxt"

		cmdReader, err := cmd.StdoutPipe()
		ErrorCheck(err, true)

		scanner := bufio.NewScanner(cmdReader)

		cmd.Start()
		ResetWD(c_wd)

		ready := sync.Mutex{}
		ready.Lock()
		read_output := func(scanner *bufio.Scanner) {
			for scanner.Scan() {
				line := scanner.Text()
				ToScreen("SimpleTxt: ")
				ToScreen(line)
				if line == "ready" {
					ready.Unlock()
				} else if line == "changed" {
					ToChangeTracker("edit\n" + "daemon_" + port + "/" + objects[id].attributes["name"] + "\n")
				} else if line == "exit" {
					is_running[id] = false
				}
			}
		}
		go read_output(scanner)
		is_running[id] = true
		ready.Lock()
		ready.Unlock()
	}
}

func InterfaceIsRunning(id int) bool {
	return is_running[id]
}

func GetAppPort(id int) string {
	conn_port_int, err := strconv.Atoi(SimpleTxt_port)
	ErrorCheck(err, true)
	conn_port_int = conn_port_int + id
	conn_port := strconv.Itoa(conn_port_int)

	return conn_port
}
