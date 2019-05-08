package main

import (
	"github.com/GO/RelaySVR"
	"log"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := RelaySVR.Start()
	if err != nil {
		log.Println(err)
	}
}
