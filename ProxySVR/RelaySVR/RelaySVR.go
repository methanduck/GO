package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
)

type Server struct {
	SVR_Addr string
	SVR_Port string
}

var pInfo *log.Logger
var pErr *log.Logger

func main() {
	pInfo = log.New(os.Stdout, "INFO : ", log.LstdFlags)
	pErr = log.New(os.Stdout, "ERR : ", log.LstdFlags)
}

func Start() error {
	Server_port := flag.String("port", "6866", "Server Port")
	Server_Addr := flag.String("addr", "127.0.0.1", "Server Addr")

	SERVER := Server{SVR_Addr: *Server_Addr, SVR_Port: *Server_port}

	Listener, err := net.Listen("tcp", SERVER.SVR_Addr+":"+SERVER.SVR_Port)
	if err != nil {
		pErr.Panic("Failed to open server (Err code : %s ", err)
	}
	defer func() {
		if err := Listener.Close(); err != nil {
			pErr.Panic("Abnormal termination while closing server")
		}
	}()

	for {
		if connection, err := Listener.Accept(); err != nil {
			pErr.Println("Failed to connect :" + connection.RemoteAddr().String())
		} else {
			go func() {
				if err := afterConnected(connection); err != nil {
					pErr.Println("Error occured (Err code : %s", err)
				}
			}()

		}

	}
}

func afterConnected(conn net.Conn) error {
	addrParse := strings.Split(conn.RemoteAddr().String(), ":")

}
