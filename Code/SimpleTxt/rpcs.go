//filename: rpcs.go
//information: created on 5th of November 2015 by Andreas Kittilsland
//Contains specific communications functions
package main

import (
	"bufio"
	"gopkg.in/qml.v1"
	"net"
	"os"
	"strconv"
	"strings"
)

//Store current view to simplify operations
var sel_view_start = -1
var sel_view_end = -1

//Create a new document
func rpc_new_file(conn net.Conn, conn_reader *bufio.Reader) {
	GuiInterface.New_file()
	if GuiInterface.Log == "New file created" {
		WriteLines(conn, "ack")
	} else {
		WriteLines(conn, "Failed to make clean document!")
	}
	return
}

//Open a given file
func rpc_open_file(conn net.Conn, conn_reader *bufio.Reader) {
	path := ReadLine(conn_reader)

	GuiInterface.Openpath = "file:" + path
	qml.Changed(GuiInterface, &GuiInterface.Openpath)

	GuiInterface.Open_file()
	if GuiInterface.Log == "File was succesfully opened!" {
		WriteLines(conn, "ack")
	} else {
		WriteLines(conn, "Failed to open file!")
	}
	return
}

//Get the subset of the text that is visisble in the remote interface
func rpc_get_text(conn net.Conn, conn_reader *bufio.Reader) {
	abs_start_pos := ReadLine(conn_reader)
	rel_end_pos := ReadLine(conn_reader)

	txtArea := GuiInterface.Root.ObjectByName("textBox")
	length := txtArea.Property("length").(int)

	start_pos, err := strconv.Atoi(abs_start_pos)
	ErrorCheck(err, true)
	rel_end_post_i, err := strconv.Atoi(rel_end_pos)
	ErrorCheck(err, true)
	end_pos := start_pos + rel_end_post_i
	if end_pos > length {
		end_pos = length
	}

	text_to_send := txtArea.Call("getFormattedText", start_pos, end_pos).(string)
	words := len(strings.Split(text_to_send, " "))

	//Aoid splitting the last word
	for i := 1; end_pos+i <= length; i++ {
		t := txtArea.Call("getFormattedText", start_pos, end_pos+i).(string)
		if words == len(strings.Split(t, " ")) {
			text_to_send = t
		} else {
			end_pos = end_pos + i - 1
			break
		}
	}

	//Avoid splitting the first word
	for i := 1; start_pos-i >= 0; i++ {
		t := txtArea.Call("getFormattedText", start_pos-i, end_pos).(string)
		if words == len(strings.Split(t, " ")) {
			text_to_send = t
		} else {
			start_pos = start_pos - i + 1
			break
		}
	}

	//Store view for simplification of operations
	sel_view_end = end_pos
	sel_view_start = start_pos

	//Send view to daemon
	read_text := text_to_send + "\n<--done-->\n"
	WriteLines(conn, read_text)

	return
}

//Save the open file
func rpc_save_file(conn net.Conn, conn_reader *bufio.Reader) {
	FileLock.Lock()

	GuiInterface.Savepath = GuiInterface.Openpath
	qml.Changed(GuiInterface, &GuiInterface.Openpath)

	GuiInterface.Save_file()
	if GuiInterface.Log == "File was succesfully saved!" {
		WriteLines(conn, "ack")
	} else {
		WriteLines(conn, "Failed to commit file!")
	}
	return
}

//Update text
func rpc_update(conn net.Conn, conn_reader *bufio.Reader) {
	diffp := ReadIntLine(conn_reader)
	equp := ReadIntLine(conn_reader)

	//Translate position from page relative to document relative
	diffp = diffp + sel_view_start
	equp = equp + sel_view_start

	docFrag := ReadDocumentFragment(conn_reader)
	txtArea := GuiInterface.Root.ObjectByName("textBox")
	txtArea.Call("remove", diffp, equp)
	docFrag = strings.Replace(docFrag, "<--newline-->", "<br/>", -1)
	txtArea.Call("insert", diffp, docFrag)
	//fmt.Println(time.Now())
	WriteLines(conn, "ack")
	return
}

//Make bold
func rpc_make_bold(conn net.Conn, conn_reader *bufio.Reader) {
	SetTextFormattingSettings(conn, conn_reader)
	GuiInterface.Make_bold()

	GuiInterface.Savepath = GuiInterface.Openpath
	qml.Changed(GuiInterface, &GuiInterface.Openpath)
	FileLock.Lock()
	GuiInterface.Save_file()
	if GuiInterface.Log == "File was succesfully saved!" {
		WriteLines(conn, "ack")
	} else {
		WriteLines(conn, "Failed to commit file!")
	}

	return
}

//Make italic
func rpc_make_italic(conn net.Conn, conn_reader *bufio.Reader) {
	SetTextFormattingSettings(conn, conn_reader)
	GuiInterface.Make_italic()

	GuiInterface.Savepath = GuiInterface.Openpath
	qml.Changed(GuiInterface, &GuiInterface.Openpath)
	FileLock.Lock()
	GuiInterface.Save_file()
	if GuiInterface.Log == "File was succesfully saved!" {
		WriteLines(conn, "ack")
	} else {
		WriteLines(conn, "Failed to commit file!")
	}

	return
}

//Make underlined
func rpc_make_underlined(conn net.Conn, conn_reader *bufio.Reader) {
	SetTextFormattingSettings(conn, conn_reader)
	GuiInterface.Make_underlined()

	GuiInterface.Savepath = GuiInterface.Openpath
	qml.Changed(GuiInterface, &GuiInterface.Openpath)
	FileLock.Lock()
	GuiInterface.Save_file()
	if GuiInterface.Log == "File was succesfully saved!" {
		WriteLines(conn, "ack")
	} else {
		WriteLines(conn, "Failed to commit file!")
	}

	return
}

//Get start and end position of selection, and set the variables we need to change the formatting
func SetTextFormattingSettings(conn net.Conn, conn_reader *bufio.Reader) {
	startPos := ReadIntLine(conn_reader)
	endPos := ReadIntLine(conn_reader)

	//Translate selected position from page relative to document relative
	startPos = startPos + sel_view_start
	endPos = endPos + sel_view_start

	txtArea := GuiInterface.Root.ObjectByName("textBox")
	sel := txtArea.Call("getFormattedText", startPos, endPos)
	pre := txtArea.Call("getFormattedText", 0, startPos)
	post := txtArea.Call("getFormattedText", endPos, txtArea.Property("length"))
	GuiInterface.Plainselectedtext = sel.(string)
	GuiInterface.Plainbeforetext = pre.(string)
	GuiInterface.Plainaftertext = post.(string)
	qml.Changed(GuiInterface, &GuiInterface.Plainselectedtext)
	qml.Changed(GuiInterface, &GuiInterface.Plainbeforetext)
	qml.Changed(GuiInterface, &GuiInterface.Plainaftertext)

	GuiInterface.Startselection = startPos
	GuiInterface.Endselection = endPos
	qml.Changed(GuiInterface, &GuiInterface.Startselection)
	qml.Changed(GuiInterface, &GuiInterface.Endselection)
}

//Close the application
func rpc_exit(conn net.Conn, conn_reader *bufio.Reader) {
	WriteLines(conn, "ack")
	os.Exit(0)
	return
}
