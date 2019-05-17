package main

import (
	"flag"
	"os/exec"
	"runtime"
)
import (
	"github.com/methanduck/GO/InteractiveSocket"
)

func main() {
	addr, _ := exec.Command("awk `END{print $1}` /etc/hosts").Output()
	if string(addr) == "" {
		addr = []byte("127.0.0.1")
	}
	Address := flag.String("addr", string(addr), "")
	runtime.GOMAXPROCS(runtime.NumCPU())

	localServer := InteractiveSocket.Window{}
	localServer.Start()
}
