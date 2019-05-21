package main

import (
	"github.com/fatih/color"
	"log"
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
	/*
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
			addrModified := addr[:len(addr)-1]*/
	if err := run("210.125.31.153", "6866"); err != nil {
		red := color.New(color.FgRed).SprintFunc()
		log.Panic(red("Stop running" + err.Error()))
	}
}

//}
func run(address string, port string) error {
	localServer := InteractiveSocket.Window{}
	return localServer.Start(address, port)
}
