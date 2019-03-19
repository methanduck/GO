package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/GO/InteractiveSocket"
	"log"
	"net"
	"os"
	"strings"
)

type Server struct {
	SVR_Addr string
	SVR_Port string
	Pinfo    *log.Logger
	PErr     *log.Logger
}

func main() {

}

func (svr Server) Start() error {
	svr.Pinfo = log.New(os.Stdout, "INFO :", log.LstdFlags)
	svr.PErr = log.New(os.Stdout, "ERR :", log.LstdFlags)
	Server_port := flag.String("port", "6866", "Server Port")
	Server_Addr := flag.String("addr", "127.0.0.1", "Server Addr")

	SERVER := Server{SVR_Addr: *Server_Addr, SVR_Port: *Server_port}

	Listener, err := net.Listen("tcp", SERVER.SVR_Addr+":"+SERVER.SVR_Port)
	if err != nil {
		svr.PErr.Panic("Failed to open server (Err code : %s ", err)
	}
	defer func() {
		if err := Listener.Close(); err != nil {
			svr.PErr.Panic("Abnormal termination while closing server")
		}
	}()

	for {
		if connection, err := Listener.Accept(); err != nil {
			svr.PErr.Println("Failed to connect :" + connection.RemoteAddr().String())
		} else {
			go func() {
				if err := afterConnected(connection, svr.PErr); err != nil {
					svr.PErr.Println("Error occured (Err code : %s", err)
				}
			}()
		}

	}
}

func afterConnected(conn net.Conn, perr *log.Logger) error {
	//Json 해석된 result struct
	result, err := InteractiveSocket.COMM_RECVJSON(conn)
	if err != nil {
		perr.Println(err)
	}

	addrParse := strings.Split(conn.RemoteAddr().String(), ":")

}
