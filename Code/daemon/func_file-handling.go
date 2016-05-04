//filename: func_file-handling.go
//information: created on 10th of October 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var std_dir = "../files/"    // The standard directory in which the application stores files belonging to each daemon
var daemon_dir = ""          //The directory in which this daemon is running
var wd_lock = ExpandedLock{} //Lets one goroutine at a time change the current working directory

//Checks that the dir we store the files in exists, else creates it
func CheckFileDirExists() {
	wd := SetWD("../files")
	_, err := ioutil.ReadDir("daemon_" + port)
	if err != nil {
		os.Mkdir("daemon_"+port, 0777)
	}
	ResetWD(wd)

	return
}

//Change the working directory
func SetWD(new_wd string) string {
	wd_lock.Lock()

	c_wd, err := os.Getwd()
	ErrorCheck(err, true, true, true)
	os.Chdir(new_wd)

	return c_wd
}

//Reset the working directory
func ResetWD(wd string) {
	os.Chdir(wd)
	wd_lock.Unlock()
}

//Create dump file
func CreateDumpFile() {
	wd := SetWD(daemon_dir)
	f, err := os.Create("dumps/dump_" + ip_self + "|" + port + ".txt")
	ErrorCheck(err, true)

	_, err = f.WriteString(time.Now().Format(time.RFC3339) + "\n")
	ErrorCheck(err, true)
	f.Close()
	ResetWD(wd)

	return
}

//Dump timings to file
func DumpTime(t time.Duration, start string, id int) {
	var dump string
	wd := SetWD(daemon_dir)

	f, err := os.OpenFile("dumps/dump_"+ip_self+"|"+port+".txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	ErrorCheck(err, true)

	if is_coord {
		dump = ip_self + ":" + port + "|"
	} else {
		dump = coord_ip + ":" + coord_port + "|"
	}
	dump = dump + start + "|" + t.String() + "|" + strconv.Itoa(id) + "\n"

	_, err = f.WriteString(dump)
	ErrorCheck(err, true)
	f.Close()
	ResetWD(wd)

	return
}

func CheckFiles() {
	//Run the change detector that checks the local fs to start the system
	wd_lock.ReadOnly()
	cmd := exec.Command("../change_tracker/change_tracker", "daemon_"+port)
	wd_lock.Editable()

	cmdWriter, err := cmd.StdinPipe()
	ErrorCheck(err, true)
	ChangeDetector_InPipe = cmdWriter

	cmdReader, err := cmd.StdoutPipe()
	ErrorCheck(err, true)

	scanner := bufio.NewScanner(cmdReader)

	err = cmd.Start()
	ErrorCheck(err, true)

	go Monitor(scanner)

	return
}

//If file exists, return path as string
func LookForFile(subdir string, name string) string {
	wd_lock.ReadOnly()
	file_path := "N/A"
	_, err := os.Stat(std_dir + subdir + name)
	if os.IsNotExist(err) {
		//File was not found in this dir
		//Look for subdirs
		entries, err := ioutil.ReadDir(std_dir + subdir)
		ErrorCheck(err, true)

		//Explore subdirs
		for _, entry := range entries {
			if entry.IsDir() {
				res := LookForFile(subdir+entry.Name()+"/", name)
				if res != "N/A" { //Break and return if we found the file
					file_path = res
					break
				}
			}
		}
	} else if err != nil {
		ErrorCheck(err, true)
	} else {
		file_path = subdir
	}
	wd_lock.Editable()
	return file_path
}

//Opens a local file
func OpenFile(id int) {
	wd_lock.Lock()
	locks[id].ReadOnly()
	file_to_open := std_dir + objects[id].file_path + "/" + objects[id].attributes["name"]
	cmd := exec.Command("open", file_to_open)
	err := cmd.Run()
	ErrorCheck(err, true, true, true)

	locks[id].Editable()
	wd_lock.Unlock()

	return
}

func ReadMetaData(attributes map[string]string, path string, obj_t string) map[string]string {
	return_to, err := os.Getwd()
	ErrorCheck(err, true)
	wd_lock.Lock()
	os.Chdir(std_dir + path)
	cmd := exec.Command("file", attributes["name"])
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	ErrorCheck(err, true, true, true)

	var attrs []string
	line, err := out.ReadString('\n')
	ErrorCheck(err, true, true, true)
	attrs_string := strings.Split(line, ":")
	attrs = strings.Split(attrs_string[1], ",")

	os.Chdir(return_to)
	wd_lock.Unlock()

	if obj_t == "video" {
		attributes["format"] = attrs[0]
		attributes["resolution"] = attrs[2]
		attributes["fps"] = attrs[3]
		attributes["codec"] = attrs_string[2]
	} else if obj_t == "image" {
		attributes["resolution"] = attrs[1]
		attributes["format"] = attrs[2]
	} else if obj_t == "text" || obj_t == "other" {
		attributes["format"] = attrs[0]
	} else if obj_t == "sound" {
		attributes["format"] = attrs[0]
		attributes["quality"] = attrs[3]
		attributes["sample"] = attrs[4]
	}
	return attributes
}

//Read output from change detector
func Monitor(scanner *bufio.Scanner) {
	//Function that finds id of object with a given path
	HandleChangedFile := func(paths ...string) {
		path := paths[0]
		var new_path string
		if len(paths) > 1 {
			new_path = paths[1]
		}
		obj_mu.ReadOnly()
		if len(paths) == 1 { //File did not change name
			for i, obj := range objects {
				fname := obj.attributes["name"]
				if obj.file_path != "" {
					fname = obj.file_path + "/" + fname
				}
				fpath := "daemon_" + port + "/" + fname
				if fpath == path {
					//Object "i" was changed, update it
					locks[i].ReadOnly()
					attributes := obj.attributes
					attributes = ReadMetaData(attributes, obj.file_path, obj.obj_type)
				retry:
					FI, err := os.Stat(std_dir + fname)
					if ErrorCheck(err, false) {
						log.Println("File was either already open and being changed, or does not exist. Retrying in 500ms...")
						time.Sleep(time.Millisecond * 500)
						goto retry //Its ugly, but solves the problem with just two lines
					} else {
						attributes["size"] = strconv.FormatInt(FI.Size(), 10) + " bytes"
						attributes["mod_date"] = FI.ModTime().String()
						locks[i].Editable()
						obj_mu.Editable()
						err = SendInstigationRequest(i, obj.version+1, attributes, obj.obj_type)
						ErrorCheck(err)
					}
					if SyncScan.IsLocked() {
						syncwd_wg.Done()
					}
					return
				}
			}
		} else { //File has changed name
			new_fname := strings.Split(new_path, "/")[1]
			for i, obj := range objects {
				fname := new_fname
				fpath_coda := obj.attributes["name"]
				if obj.file_path != "" {
					fname = obj.file_path + "/" + new_fname
					fpath_coda = obj.file_path + "/" + fpath_coda
				}
				fpath := "daemon_" + port + "/" + fpath_coda
				if fpath == path {
					//Object "i" was changed, update it
					locks[i].ReadOnly()
					attributes := obj.attributes
					attributes["name"] = new_fname
					attributes = ReadMetaData(attributes, obj.file_path, obj.obj_type)
					FI, err := os.Stat(std_dir + fname)
					ErrorCheck(err, true)
					attributes["size"] = strconv.FormatInt(FI.Size(), 10) + " bytes"
					attributes["mod_date"] = FI.ModTime().String()
					locks[i].Editable()
					obj_mu.Editable()
					err = SendInstigationRequest(i, obj.version+1, attributes, obj.obj_type)
					ErrorCheck(err)
					if SyncScan.IsLocked() {
						syncwd_wg.Done()
					}
					return
				}
			}
		}
		obj_mu.Editable()
		if SyncScan.IsLocked() {
			syncwd_wg.Done()
		}
	}

	HandleAddedFile := func(path string) {
		ext_t := strings.Split(path, ".")
		ext := ext_t[len(ext_t)-1]
		obj_type := "other"
		if ext == "txt" || ext == "stxt" {
			obj_type = "text"
		} else if ext == "wav" {
			obj_type = "sound"
		} else if ext == "avi" {
			obj_type = "video"
		} else if ext == "png" {
			obj_type = "image"
		}

		FI, err := os.Stat(std_dir + path[11:len(path)])
		ErrorCheck(err, true)

		attributes := make(map[string]string)
		attributes["name"] = FI.Name()
		attributes["mod_date"] = FI.ModTime().String()
		attributes["size"] = strconv.FormatInt(FI.Size(), 10) + " bytes"

		err = SendInstigationRequest(-1, 1, attributes, obj_type)
		if SyncScan.IsLocked() {
			syncwd_wg.Done()
		}
	}

	HandleDeletedFile := func(path string) {
		obj_mu.ReadOnly()
		for i, obj := range objects {
			fname := obj.attributes["name"]
			if obj.file_path != "" {
				fname = obj.file_path + "/" + fname
			}
			fpath := "daemon_" + port + "/" + fname
			if fpath == path {
				obj_mu.Editable()
				err := SendDeletionRequest(i)
				ErrorCheck(err)
				if SyncScan.IsLocked() {
					syncwd_wg.Done()
				}
				return
			}
		}
		obj_mu.Editable()
		if SyncScan.IsLocked() {
			syncwd_wg.Done()
		}
	}

	//Keep in touch with the change tracker
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			if line[0:1] == "!" {
				fmt.Println(line)
			} else if line == "sync run done" {
				//For the syncronous change tracker scan. Tells us when it is done
				SyncScan.Unlock()
			} else {
				info := strings.Split(line, ":")
				if SyncScan.IsLocked() {
					syncwd_wg.Add(1)
				}
				if len(info) > 1 {
					if info[0] == "add" {
						go HandleAddedFile(info[1])
					} else if info[0] == "delete" {
						go HandleDeletedFile(info[1])
					} else if info[0] == "PID" {
						chg_trk_pid_string = info[1]
					} else {
						go HandleChangedFile(info[0], info[1])
					}
				} else {
					go HandleChangedFile(line)
				}
			}
		}
	}

	return
}

var SyncScan = ExpandedLock{}
var syncwd_wg sync.WaitGroup

func RunChangeTrackerScan(async bool) {
	if !is_coord {
		FindApplications()
	} else {
		coord_FindApplications()
	}
	if async {
		if is_coord {
			SendToAllDaemons("run chng dtctr\nasync\n")
		}
		ToChangeTracker("run\n")
	} else {
		SyncScan.Lock()
		if is_coord {
			SendToAllDaemons("run chng dtctr\nsync\n")
		}
		ToChangeTracker("sync run\n")
		SyncScan.Lock()
		syncwd_wg.Wait()
		SyncScan.Unlock()
	}
	return
}
