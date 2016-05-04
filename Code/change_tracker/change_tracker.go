//filename: watchdog.go
//information: created on 26th of November 2015 by Andreas Kittilsland
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) == 2 {
		//Change working directory to the root folder
		root_folder = os.Args[1]
		err := os.Chdir("../files")
		ErrorCheck(err, true)

	} else {
		ErrorCheck(errors.New("Number of arguments were "+strconv.Itoa(len(os.Args)-1)+". Need root directory to walk as argument."), true)
	}
	GetPID()

	//Map of the objects/files we find in the folder
	objects = make(map[string]os.FileInfo)

	//Do initial walk of the folder to construct a map of the files
	wlkfnc := func(path string, info os.FileInfo, err error) error {
		t := strings.Split(info.Name(), ".")
		ext := t[len(t)-1]
		if ErrorCheck(err) {
			return err
		} else if info.Name() != ".DS_Store" && ext != "sh" {
			objects[path] = info
			if !info.IsDir() {
				AddFile(path)
			}
		}
		return nil
	}
	err := filepath.Walk(root_folder, wlkfnc)
	ErrorCheck(err)

	//for true {
	DoWalk(true)
	//}
	ReadStdin()
}

//Walk the folder continously to look for changes
func DoWalk(async bool) {
	//Map over the files we found on this pass
	this_pass := make(map[string]bool)
	possible_name_changes := make(map[string]string)
	wf := func(path string, info os.FileInfo, err error) error {
		t := strings.Split(info.Name(), ".")
		ext := t[len(t)-1]
		if ErrorCheck(err) {
			return err
		} else if info.Name() != ".DS_Store" && ext != "sh" {
			//Check it off that we saw the file on this pass
			this_pass[path] = true

			//Check if the file exists in our map
			if _, ok := objects[path]; !ok {
				//It was not in the map, try to see if it just changed name
				for i, obj := range objects {
					if obj.ModTime() == info.ModTime() && obj.Size() == info.Size() {
						objects[path] = info
						//Add it to list of possible name changes to go through before deletion at end of walk
						possible_name_changes[i] = path
						return nil
					}
				}

				//If we didnt find a candidate, add it to the list
				objects[path] = info
				AddFile(path)
			} else {
				//Check if the file has changed since last walk
				if info.ModTime() != objects[path].ModTime() {
					//Only tell the daemon about the change if it is not a folder
					if !info.IsDir() {
						ChangeFile(path)
					}
					objects[path] = info
				}
			}
		}
		return nil
	}
	err := filepath.Walk(root_folder, wf)
	ErrorCheck(err)

	//Delete files that were no longer found in their position
	for old_path, _ := range objects {
		if _, ok := this_pass[old_path]; !ok {
			if new_path, found := possible_name_changes[old_path]; found {
				//Probability says it is the same file with a new name
				//Lets simply tell the daemon that the deleted file changed name to the added file
				delete(objects, old_path)
				delete(possible_name_changes, old_path)
				NameChangeFile(old_path, new_path)
			} else {
				delete(objects, old_path)
				DeleteFile(old_path)
			}
		}
	}

	//Tell the daemon to add the remaining new files that were not found to be name changes
	for _, path := range possible_name_changes {
		AddFile(path)
	}

	if !async {
		FinishSyncedScan()
	}
}
