package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"
)
import (
	"github.com/methanduck/GO/InteractiveSocket"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	addr, _ := exec.Command("sh", "-c", "awk 'END{print $1}' /etc/hosts").Output()
	if string(addr) == "" {
		addr = []byte("127.0.0.1")
	}
	address := flag.String("addr", string(addr), "Set listening address")
	port := flag.String("port", InteractiveSocket.SVRLISTENINGPORT, "Set listening port")
	flag.Parse()

	fmt.Println(*address + *port)
	localServer := InteractiveSocket.Window{}
	localServer.Start(*address, *port)
}
