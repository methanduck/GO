package main

import (
	"flag"
	"github.com/methanduck/GO/RelaySVR"
	"log"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Server_port := flag.String("port", RelaySVR.Service_port, "Server Port")
	Server_Addr := flag.String("addr", "", "Server Addr")

	err := RelaySVR.Start(*Server_Addr, *Server_port)
	if err != nil {
		log.Println(err)
	}
}
