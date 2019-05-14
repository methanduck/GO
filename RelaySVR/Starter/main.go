package main

import (
	"github.com/methanduck/GO/RelaySVR"
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
