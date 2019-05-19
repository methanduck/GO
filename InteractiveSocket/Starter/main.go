package main

import (
	"os"
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
	port := os.Getenv("windowprt")
	localServer := InteractiveSocket.Window{}
	localServer.Start(address, port)
}
