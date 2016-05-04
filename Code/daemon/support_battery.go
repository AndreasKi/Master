//filename: support_battery.go
//information: created on 13th of January 2016 by Andreas Kittilsland
package main

import (
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

//Checks status on battery, returns (true, x%) if it is being charged, and (false, x%) if on battery.
func CheckBattery() (bool, int) {
	is_charging := false
	percentage_remaining := 0
	if runtime.GOOS == "darwin" { //OSX
		var cmdOut []byte
		cmdOut, err := exec.Command("pmset", "-g", "batt").Output()
		if err != nil { //Likely running on a desktop
			percentage_remaining = 100
			is_charging = true
		} else {
			//Output example:
			//Now drawing from 'AC Power'
			//-InternalBattery-0	100%; charged; 0:00 remaining present: true
			output_line := string(cmdOut)
			unformatted_o := strings.Split(output_line, "\n") //["Now drawing from 'AC Power'","-InternalBattery-0	100%; charged; 0:00 remaining present: true"]

			//Whether or not it is currently charging the battery
			charging_status := strings.Split(unformatted_o[0], "'")[1] //"AC Power"

			//What percentage the battery is currently at
			temp_o := strings.Split(unformatted_o[1], ";")[0] //"-InternalBattery-0	100%"
			percentage := strings.Split(temp_o, "\t")[1]      //"100%"

			percentage_remaining, err = strconv.Atoi(percentage[:len(percentage)-1])
			ErrorCheck(err, true)

			if charging_status == "AC Power" {
				is_charging = true
			}
		}
	} else if runtime.GOOS == "linux" { //Linux
		var cmdOut []byte
		cmdOut, err := exec.Command("upower", "-i", "/org/freedesktop/UPower/devices/battery_BAT0|", "grep", "-E", "\"state|percentage\"").Output()
		if err != nil { //Likely running on a desktop
			percentage_remaining = 100
			is_charging = true
		} else {
			//Output example:
			//state:               charging
			//percentage:          100%
			output_line := string(cmdOut)
			unformatted_o := strings.Split(output_line, "\n") //["state:	charging", "percentage:		100%"]

			//Whether or not it is currently charging the battery
			charging_status := strings.Split(unformatted_o[0], "\t")[1] //"charging"

			//What percentage the battery is currently at
			percentage := strings.Split(unformatted_o[1], "\t")[1] //"100%"

			percentage_remaining, err = strconv.Atoi(percentage[:len(percentage)-1])
			ErrorCheck(err, true)

			if charging_status == "charging" {
				is_charging = true
			}
		}
	} else { //Windows or FreeBSD
		ErrorCheck(errors.New("Unable to check battery status as OS is not recognized."), true)
	}
	return is_charging, percentage_remaining
}
