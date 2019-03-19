package RelaySVR

import (
	"flag"
	"github.com/GO/InteractiveSocket"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type Server struct {
	State    *dbData
	SVR_Addr string
	SVR_Port string
	Pinfo    *log.Logger
	PErr     *log.Logger
}

func Start() error {
	Server_port := flag.String("port", "6866", "Server Port")
	Server_Addr := flag.String("addr", "127.0.0.1", "Server Addr")
	SERVER := Server{SVR_Addr: *Server_Addr, SVR_Port: *Server_port}
	SERVER.Pinfo = log.New(os.Stdout, "INFO :", log.LstdFlags)
	SERVER.PErr = log.New(os.Stdout, "ERR :", log.LstdFlags)
	//bolt database initializing
	SERVER.State = new(dbData)
	SERVER.State.Startbolt(SERVER.Pinfo, SERVER.PErr)
	Listener, err := net.Listen("tcp", SERVER.SVR_Addr+":"+SERVER.SVR_Port)
	if err != nil {
		SERVER.PErr.Panic("Failed to open server (Err code : %s ", err)
	}
	defer func() {
		if err := Listener.Close(); err != nil {
			SERVER.PErr.Panic("Abnormal termination while closing server")
		}
	}()

	for {
		if connection, err := Listener.Accept(); err != nil {
			SERVER.PErr.Println("Failed to connect :" + connection.RemoteAddr().String())
		} else {
			//수신 시
			go func() {
				if err := SERVER.afterConnected(connection, SERVER.PErr); err != nil {
					SERVER.PErr.Println("Error occured (Err code : %s", err)
				}
			}()
		}

	}
}

func (server Server) afterConnected(conn net.Conn, perr *log.Logger) error {
	//Json 해석된 result struct
	result, err := InteractiveSocket.COMM_RECVJSON(conn)
	if err != nil {
		perr.Println(err)
	}
	switch result.Which {
	//Application
	case true:
		status, err := server.State.IsOnline(result.Identity)
		if err != nil {
			server.Pinfo.Println("Send Ack : ERR")
			if err := InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: InteractiveSocket.COMM_ERR}, conn); err != nil {
				server.PErr.Println("Failed to send JSON")
			}
		}
		if status == OFFLINE {
			_ = InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: InteractiveSocket.COMM_FAIL}, conn) //ERR처리 무시함
		} else {
			//ONLINE일 경우
			waiting := 1
			if err := server.State.ApplicationNodeUpdate(result, waiting); err != nil {
				if err.Error() == NOTFOUND {
					_ = InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: InteractiveSocket.COMM_ERR}, conn) //ERR처리 무시함
				}
			}
			for {
				if waiting == 3 {

				}
			}

		}

	//Window
	case false:

	}
	addrParse := strings.Split(conn.RemoteAddr().String(), ":")

}
