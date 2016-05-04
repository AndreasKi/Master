//filename: struct_object.go
//information: created on 10th of March 2016 by Andreas Kittilsland
package main

/*
attributes is a map using a string as a key, holding all values as a string
The keys used depend on the type of file we are holding information on
*/
type object struct { //Structure for holding the metadata of the objects
	version    int
	obj_type   string //string saying what type of file this object is {text, sound, video, image, other}
	attributes map[string]string
	file_path  string
	extension  string
}

var obj_mu = ExpandedLock{}     //Mutex for the list of objects in our system (lock.go)
var objects map[int]object      //List of objects in our system. Index is the object ID
var locks map[int]*ExpandedLock //List of locks for each object (lock.go)

func (obj object) SetFilePath(path string) {
	obj.file_path = path
	return
}
