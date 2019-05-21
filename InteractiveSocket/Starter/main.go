package main

import (
	"github.com/fatih/color"
	"log"
	"os"
	"os/exec"
	"runtime"
)
import (
	"github.com/methanduck/GO/InteractiveSocket"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	address := os.Getenv("windowaddr")
	port := os.Getenv("windowport")

	if address == "" {
		addr, err := exec.Command("/bin/sh", "-c", "awk 'END{print $1}' /etc/hosts").Output()
		if err != nil {
			color.Set(color.FgRed)
			defer color.Unset()
			log.Println("ERROR!! Socket server failed to get address or port!!")
			log.Panic("Aborting initialize" + err.Error())
		}
		addrModified := addr[:len(addr)-1]
		if err := run(string(addrModified), port); err != nil {
			red := color.New(color.FgRed).SprintFunc()
			log.Panic(red("Stop running" + err.Error()))
		}
	}

}
func run(address string, port string) error {
	localServer := InteractiveSocket.Window{}
	return localServer.Start(address, port)
}
