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
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := Start()
	log.Println(err)
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

func (server Server) afterConnected(conn net.Conn, perr *log.Logger) {
	//Json 해석된 result struct
	result, err := InteractiveSocket.COMM_RECVJSON(conn)
	if err != nil {
		perr.Println(err)
	}
	//들어온 데이터를 창문과 어플리케이션으로 구별 처리
	switch result.Which {
	//Application
	/*1. Node 구조체가 수신되면 해당 구조체의 Identity 를 참고, 창문이 현재 서버에서 온라인 상태인지 확인함
	  2-1. 온라인 상태가 아닐경우 :  "ERR" string값을 어플리케이션에 전송함
	  2-2. 온라인 상태일 경우 : 데이터를 리턴받는 명령이 아닐경우 Node구조체를 창문이 가져갈 수 있도록 데이터베이스에 올려둠 , 데이터를 리턴받는 경우 창문이 재 접속하기 위한 시간(`2sec)보다 큰 3초를 대기 후
							  데이터베이스에서 창문이 올려둔 데이터를 수령해 간다.
	*/
	case true:
		//창문의 온라인 여부를 데이터베이스에 쿼리
		status, err := server.State.IsExistAndIsOnline(result.Identity)
		if err != nil { //데이터베이스에 존재하지 않거나 처리중 오류 발생 시 오류 내용이 어플리케이션에 전달됨
			server.Pinfo.Println("Send Ack : ERR")
			if err := InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: InteractiveSocket.COMM_ERR}, conn); err != nil {
				server.PErr.Println("Failed to send JSON")
			}
		}
		//리턴값이 오류없이 존재한다면 데이터베이스에 현재 창문이 존재함을 의미
		if !status { //서버에 창문이 OFFLINE으로 확인 될 경우
			_ = InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: InteractiveSocket.STATE_OFFLINE}, conn)
		} else { //서버에 창문이 ONLINE으로 확인 될 경우
			//어플에서 수신된 데이터를 데이터베이스의 창문과 연관지어 저장 후 창문이 가져 갈 수 있게 함
			// isRequireConn : 창문이 현재 대기중인 데이터가 존재함을 알 수 있게 함 , lock : 이미 명령이 대기중이므로 다른 어플리케이션에서 추가 명령을 넣을 수 없게 함
			if err := server.State.UpdateNodeDataState(result, false, true, 1, UPDATE_REQCONN); err != nil {
				//노드 업데이트 오류 발생 시 어플리케이션에 전송함
				perr.Println(err)
				_ = InteractiveSocket.COMM_SENDJSON(&InteractiveSocket.Node{Ack: err.Error()}, conn) //TODO : 오류 종류에 대한 처리 없이 오류 사항을 그대로 전송중
			}
			//수신된 Node에서 명령 여부에 따라 응답을 대기하거나 전송하고 종료함
			switch result.Oper {
			case "INFO": //TODO : 재수정 필요
				time.Sleep(3 * time.Second)
				if window, err := server.State.GetNodeData(result.Identity); err != nil {
					perr.Println(err)
				} else {
					window.ApplicationData.Ack = InteractiveSocket.COMM_SUCCESS
					_ = InteractiveSocket.COMM_SENDJSON(&window.ApplicationData, conn)
					if err := server.State.UpdateNodeDataState(InteractiveSocket.Node{}, true, false, 0, UPDATE_ALL); err != nil {
						perr.Println(err)
					}
				}
			default:
				//기본으로 명령은 전송으로 종료함을 원칙으로 하고 어플리케이션과의 접속을 종료함
				return
			}
		}
	//Window
	//창문의 경우 한번이라도 신호를 보내오면 온라인 연결 간주, 대기중인 명령이 있는지 확인 후 명령 처리 및 응답
	case false:
		//데이터베이스에 해당 창문을 온라인 상태로 업데이트함
		if err := server.State.UpdataOnline(result); err != nil {
			perr.Println(err)
		}
		//창문에 대해 대기중인 요청이 있는지 데이터베이스에서 확인
		if reqConn, err := server.State.IsRequireConn(result.Identity); err != nil {
			perr.Println(err.Error())
			return
		} else {
			if reqConn {
				if nodeData, err := server.State.GetNodeData(result.Identity); err != nil {
					_ = InteractiveSocket.COMM_SENDJSON(&nodeData.ApplicationData, conn)
				}

			}
		}
	}
}
