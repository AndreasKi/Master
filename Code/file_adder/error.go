//filename: error.go
//information: created on 30th of August 2015 by Andreas Kittilsland
package main

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strconv"
)

//Checks and handles errors. Outputs error, and closes application if error is fatal.
//optional_args[0] = Whether or not the error is fatal
//optional_args[1] = Whether or not to show error message
//optional_args[2] = Whether or not the error call was made from within a function called several places in our code
//(We want to output error at a call step higher than usual if that is the case)
func ErrorCheck(err error, optional_args ...bool) bool {
	return_val := false
	if err != nil {
		return_val = true
		show_error := true
		nested_error := false
		fatal := false
		if len(optional_args) > 0 {
			fatal = optional_args[0]
			if len(optional_args) > 1 {
				show_error = optional_args[1]
				if len(optional_args) > 2 {
					nested_error = optional_args[2]
				}
			}
		}

		caller_func := 1
		if nested_error {
			caller_func = 2
		}

		pc, _, line, _ := runtime.Caller(caller_func)
		// Retrieve a Function object this functions parent
		functionObject := runtime.FuncForPC(pc)
		// Regex to extract just the function name (and not the module path)
		extractFnName := regexp.MustCompile(`^.*\.(.*)$`)
		name := extractFnName.ReplaceAllString(functionObject.Name(), "$1")

		fmt.Print("\n")
		if fatal {
			log.Fatal("Function "+name+" on line "+strconv.Itoa(line)+" - Failed fatally: ", err)
		} else if show_error {
			log.Println("Function "+name+" - Had error: ", err)
		}
	}
	return return_val
}
