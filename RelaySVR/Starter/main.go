package main

import (
	"github.com/methanduck/GO/RelaySVR"
	"log"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	/* change flag to os.Getenv
	Server_port := flag.String("port", RelaySVR.Service_port, "Server Port")
	Server_Addr := flag.String("addr", "", "Server Addr")
	*/
	address := os.Getenv("serverwindow")
	port := os.Getenv("serverport")
	err := RelaySVR.Start(address, port)
	if err != nil {
		log.Println(err)
	}
}
