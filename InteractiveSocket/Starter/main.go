package main

import "runtime"
import (
	"github.com/GO/InteractiveSocket"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	localServer := InteractiveSocket.Window{}
	localServer.Start()
}
