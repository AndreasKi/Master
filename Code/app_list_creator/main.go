package main

import (
	"io/ioutil"
	"time"
)

const exec_time = time.Second * 6

func main() {
	t_start := time.Now()

	dat, err := ioutil.ReadFile("list.txt")
	if err != nil {
    	HandleError(err)
	}
    list := string(dat)

    time.Sleep(exec_time - time.Since(t_start))

    ReturnList(list)

    return
}
