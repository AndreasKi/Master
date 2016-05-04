//filename: header.go
//information: created on 26th of November 2015 by Andreas Kittilsland
//Contains global variables and structures

package main

import (
	"os"
)

//Number of milliseconds to rest between each walk
const WalkInterval = 500

var root_folder string
var objects map[string]os.FileInfo
