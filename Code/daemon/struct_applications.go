//filename: struct_applications.go
//information: created on 6th of December 2015 by Andreas Kittilsland
package main

import (
	"bufio"
	"os/exec"
	"strings"
	"errors"
)

type application struct {
	app_name string
	count    int
}

var sendApplicationsTo_mu = ExpandedLock{}
var sendApplicationsTo []daemon //List of daemons that need the list of applications
var applications_mu = ExpandedLock{}
var applications []application //List of all applications
var local_apps_mu = ExpandedLock{}
var local_apps []string //list of local apps

//Function that checks if an app is already in a slice list of apps
func AppIn_Struct(list []application, app string) (int, bool) {
	for i, c_app := range list {
		if c_app.app_name == app {
			return i, true
		}
	}
	return -1, false
}

//Function that checks if an app is already in a slice list of apps
func AppIn_String(list []string, app string) (int, bool) {
	for i, c_app := range list {
		if c_app == app {
			return i, true
		}
	}
	return -1, false
}

func RetrieveLocalApplications() (bool, []string, []string) {
	if !test_applications {
		changed := false
		previous_list := local_apps

		cmd := exec.Command("system_profiler", "SPApplicationsDataType")

		cmdReader, err := cmd.StdoutPipe()
		ErrorCheck(err, true)

		scanner := bufio.NewScanner(cmdReader)

		err = cmd.Start()
		ErrorCheck(err, true)

		var appInfo []string
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.Trim(line, ": ")
			appInfo = append(appInfo, line)
		}

		var new_apps []string
		for i, app := range appInfo {
			strtPnt := i - 1
			stpPnt := i + 1
			if strtPnt == -1 {
				strtPnt = 0
			}
			if stpPnt == len(appInfo) {
				stpPnt--
			}
			if appInfo[strtPnt] == "" && appInfo[stpPnt] == "" {
				if _, ok := AppIn_String(local_apps, app); !ok && app != "" && app != " " && app != "\n" {
					changed = true
					new_apps = append(new_apps, app)
					local_apps = append(local_apps, app)
				}
			}
		}

		return changed, previous_list, new_apps
	} else {
		changed := false
		previous_list := local_apps

		cmd := exec.Command("../app_list_creator/app_list_creator")
		cmd.Dir = daemon_dir[:len(daemon_dir)-6] + "app_list_creator"

		cmdReader, err := cmd.StdoutPipe()
		ErrorCheck(err, true)

		scanner := bufio.NewScanner(cmdReader)

		err = cmd.Start()
		ErrorCheck(err, true)

		var appInfo []string
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) > 6 && line[:6] == "Error:" {
				ErrorCheck(errors.New(line), true)
			}
			appInfo = append(appInfo, line)
		}

		var new_apps []string
		for _, app := range appInfo {
			if _, ok := AppIn_String(local_apps, app); !ok && app != "" && app != " " && app != "\n" {
				changed = true
				new_apps = append(new_apps, app)
				local_apps = append(local_apps, app)
			}
		}

		return changed, previous_list, new_apps
	}	
}
