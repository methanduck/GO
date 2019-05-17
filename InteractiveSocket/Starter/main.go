package main

import "runtime"
import (
	"github.com/methanduck/GO/InteractiveSocket"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	localServer := InteractiveSocket.Window{}
	localServer.Start()
}
