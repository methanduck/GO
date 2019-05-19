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

	/* change flag to os.Getenv
	addr, _ := exec.Command("sh", "-c", "awk 'END{print $1}' /etc/hosts").Output()
	if string(addr) == "" {
		addr = []byte("127.0.0.1")
	}
	address := flag.String("addr", string(addr), "Set listening address")
	port := flag.String("port", InteractiveSocket.SVRLISTENINGPORT, "Set listening port")
	flag.Parse()
	*/

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
		run(string(addr), port)
	}

}
func run(address string, port string) error {
	localServer := InteractiveSocket.Window{}
	localServer.Start(address, port)
	return nil
}
