//filename: func_mutators.go
//information: created on 8th of September 2015 by Andreas Kittilsland
package main

import (
	"errors"
	"strconv"
	"strings"
	//"time"
)

func FindFreeID() int {
	id := -1
	if _, exists := objects[len(objects)]; !exists {
		id = len(objects)
	} else {
		found_objs := 0
		for i := 0; found_objs < len(objects); i++ {
			if _, exists := objects[i]; !exists {
				id = i
				break
			} else {
				found_objs++
			}
		}
	}

	return id
}

//Add the object
func AddObject(id int, version int, attributes map[string]string, timer string, obj_type string) (int, error) {
	obj_mu.Lock()
	ver_vector_mu.Lock()
	//Create ID for the new object if it is brand new
	return_id := id
	if id == -1 {
		return_id = FindFreeID()
		id = return_id
	}

	//Find file extension
	t := strings.Split(attributes["name"], ".")
	ext := ""
	if len(t) > 1 {
		ext = t[1]
	}

	if _, found := objects[id]; !found { //Its a new object. Add it
		path := LookForFile("", attributes["name"])
		objects[id] = object{version, obj_type, attributes, path, ext} //LookForFile in file_handling.go
		if path == "N/A" {
			//CreateShortcut(id)
		} else {
			nr_local_objs++
		}
	} else {
		if objects[id].version < version { //Object exists, but is old. Update it
			delete(objects, id)

			//make the object
			path := LookForFile("", attributes["name"])
			objects[id] = object{version, obj_type, attributes, path, ext}
			if path == "N/A" {
				//CreateShortcut(id)
			}
		} else {
			//If they are completely equal, dont bother throwing error. It is not a conflict.
			//Overlap might have occured on reconnect if more than one daemon differs from the coordinator.
			//Both might have tried to send the same object at the same time
			sync_err_occured := true
			if objects[id].version == version && objects[id].obj_type == obj_type {
				equal := true
				for index, item := range objects[id].attributes {
					if item != attributes[index] {
						equal = false
						break
					}
				}
				if equal {
					sync_err_occured = false
				}
			}
			if sync_err_occured {
				obj_mu.Unlock()
				ver_vector_mu.Unlock()
				return return_id, errors.New("A synchronization conflict has occured! Attempted to overwrite id " +
					strconv.Itoa(id) + " of version " + strconv.Itoa(objects[id].version) + " with " + strconv.Itoa(version))
			}
		}
	}
	//t1, err := time.Parse(time.RFC3339, timer)
	//ErrorCheck(err, true)
	//t_dur := time.Since(t1)
	ToScreen("> Object added or changed. Now has "+strconv.Itoa(len(objects))+" data files.")
	//ToScreen("> Object added or changed. Now has "+strconv.Itoa(len(objects))+" data files."+
	//	"\n> Time from instigation to local synchronization: ", t_dur)

	//DumpTime(t_dur, timer, id)

	version_vector[id] = version
	ver_vector_mu.Unlock()

	obj_mu.Unlock()

	//Check to see if we are synced now
	if !is_coord && initialize.IsLocked() {
		SyncUpdate()
	}

	return return_id, nil
}

//Remove object from the network
func RemoveObject(id int, timer string) error {
	var outcome error
	outcome = nil
	obj_mu.Lock()
	ver_vector_mu.Lock()
	if _, found := objects[id]; found {
		if objects[id].file_path == "N/A" {
			//DeleteShortcut(id)
		} else {
			nr_local_objs--
		}
		delete(objects, id)
		delete(locks, id)
		delete(version_vector, id)
		//t1, err := time.Parse(time.RFC3339, timer)
		//ErrorCheck(err, true)
		//t_dur := time.Since(t1)
		ToScreen("> Object removed. Now has "+strconv.Itoa(len(objects))+" data files.")
		//ToScreen("> Object removed. Now has "+strconv.Itoa(len(objects))+" data files."+
		//	"\n> Time from instigation to local synchronization: ", t_dur)
		//DumpTime(t_dur, timer, id)

	} else {
		outcome = errors.New("Object was not found")
	}
	ver_vector_mu.Unlock()
	obj_mu.Unlock()

	//Check to see if we are synced now
	if !is_coord && initialize.IsLocked() {
		SyncUpdate()
	}

	return outcome
}

//Runs whenever a change is made and still in the initialization phase. Checks if daemon is syned with coordinator
func SyncUpdate() {
	result := true
	coord_ver_vector_mu.ReadOnly()
	ver_vector_mu.ReadOnly()
	//Check from coordinators POV
	for index, version := range coord_version_vector {
		if local_v, ok := version_vector[index]; ok { //If the current index of coord. v. vector exists in local v. vector
			if version != local_v { //Check to see if they have the same version
				result = false
				break
			}
		} else {
			result = false
			break
		}
	}
	//Check from local POV
	for index, version := range version_vector {
		if coord_v, ok := coord_version_vector[index]; ok { //If the current index of local v. vect. exists in coord v. vect.
			if version != coord_v { //Check to see if they have the same version
				result = false
				break
			}
		} else {
			result = false
			break
		}
	}
	coord_ver_vector_mu.Editable()
	ver_vector_mu.Editable()
	if result == true {
		initialize.Unlock()
		ToScreen("> Synchronization complete. Local files may now be changed.")
	}
}

//Add the daemon to the network
func AddDaemon(ip_d string, port_d string) bool {
	//Lets not add ourself
	if ip_d == ip_to && port_d == port {
		return false
	} else {
		daemons_mu.Lock()
		//Check the daemons slice. No point adding it if it is already there
		new_daemon := daemon{ip_d, port_d}
		in_daemons := false
		for _, i := range daemons {
			if i.ip == new_daemon.ip && i.port == new_daemon.port {
				in_daemons = true
				break
			}
		}

		if !in_daemons {
			//Append it to the slice of daemons
			daemons = append(daemons, new_daemon)
			nrDs := strconv.Itoa(len(daemons)+2)
			if is_coord {
				nrDs = strconv.Itoa(len(daemons)+1)
			}
			ToScreen("> Daemon added. Now has " + nrDs + " daemons in network.")
		}
		
		daemons_mu.Unlock()
		return !in_daemons
	}
}

//Remove daemon from the network
func RemoveDaemon(index int) {
	not_locked := false
	if !daemons_mu.IsLocked() {
		/*We assumse that if this function is called, it should always proceed.
		However, to avoid changes done to the slice while a daemon is being removed, we ensure it is locked*/
		daemons_mu.Lock()
		not_locked = true
	}
	daemons[index] = daemons[len(daemons)-1]
	daemons = daemons[:len(daemons)-1]
	nrDs := strconv.Itoa(len(daemons)+2)
	if is_coord {
		nrDs = strconv.Itoa(len(daemons)+1)			
	}
	ToScreen("> Daemon removed. Now has " + nrDs + " daemons in network.")
	if not_locked {
		daemons_mu.Unlock()
	}

	return
}
