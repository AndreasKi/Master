package main

import (
	"fmt"
	"os"
)

func HandleError(err error) {
	fmt.Println("Error: ", err)
	os.Exit(1)
}

func ReturnList(list string) {
	fmt.Println(list)
	return
}
