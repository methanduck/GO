/*
	Main Relay Server
*/
package RelaySVR

import (
	"flag"
	"github.com/GO/InteractiveSocket"
	"log"
	"net"
	"os"
	"runtime"
	"time"
)

const (
	Service_port = "6866"
)

type Server struct {
	State    *dbData
	SVR_Addr string
	SVR_Port string
	Pinfo    *log.Logger
	PErr     *log.Logger
	// ctx		 context.Context TODO Context 추가 여부 검토
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := Start()
	if err != nil {
		log.Println(err)
	}
}

//Start Serer
func Start() error {
	Server_port := flag.String("port", Service_port, "Server Port")
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

	go func() {

	}()

	for {
		if connection, err := Listener.Accept(); err != nil {
			SERVER.PErr.Println("Failed to connect :" + connection.RemoteAddr().String())
		} else {
			//수신 시
			go func() {
				SERVER.afterConnected(connection, SERVER.PErr)
			}()
		}

	}
}

//통신이 수립되었을 때 수행하는 함수
func (server Server) afterConnected(conn net.Conn, perr *log.Logger) {

	//Json 해석된 result struct
	result, err := InteractiveSocket.COMM_RECVJSON(conn)
	if err != nil {
		perr.Println(err)
	}
	switch result.Which {
	//Application
	case true:
		status, err := server.State.IsExistAndIsOnline(result.Identity)
		if err != nil {
			server.Pinfo.Println("Send Ack : ERR")
			if err := InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: InteractiveSocket.COMM_ERR}, conn); err != nil {
				server.PErr.Println("Failed to send JSON")
			}
		}
		if !status { //서버에서 offline일 경우 조종이 불가하여 offline응답을 전송
			_ = InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: InteractiveSocket.STATE_OFFLINE}, conn)
		} else { //online확인
			if err := server.State.UpdateNodeDataState(result, false, true, 1, UPDATE_REQCONN); err != nil {
				perr.Println(err)
				_ = InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: err.Error()}, conn) //TODO : 오류 종류에 대한 처리 없이 오류 사항을 그대로 전송중
			}

			switch result.Oper {
			case "INFO": //TODO : 재수정 필요
				time.Sleep(3 * time.Second)
				if window, err := server.State.GetNodeData(result.Identity); err != nil {
					perr.Println(err)
				} else {
					window.ApplicationData.Ack = InteractiveSocket.COMM_SUCCESS
					_ = InteractiveSocket.COMM_SENDJSON(&window.ApplicationData, conn)
					if err := server.State.ResetState(result.Identity, true, false, 0); err != nil {
						perr.Println(err)
					}
				}
			default:

			}
		}
	//Window
	//창문의 경우 한번이라도 신호를 보내오면 온라인 연결 간주, 대기중인 명령이 있는지 확인 후 명령 처리 및 응답
	case false:
		switch result.Oper {
		case "INFO":
			if err := server.State.UpdateNodeDataState(result, true, false, 1, UPDATE_ALL); err != nil {
				perr.Println(err)
			}
		case "ONLINE": //주기적 수신
			if err := server.State.UpdataOnline(result); err != nil {
				perr.Println(err)
			}
		}
		if err := server.State.UpdataOnline(result); err != nil {
			perr.Println(err)
		}
	}
}
