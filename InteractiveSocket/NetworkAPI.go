package InteractiveSocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//고정 변수
const (
	//서버 포트
	SVRLISTENINGPORT = "6866"
	//중계서버 IP
	RELAYSVRIPADDR = "127.0.0.1:6866"
)

type Window struct {
	PInfo      *log.Logger
	PErr       *log.Logger
	svrInfo    *Node
	Available  *sync.Mutex
	FAvailable *sync.Mutex
	quitSIGNAL chan os.Signal
}

//VALIDATION 성공 : "LOGEDIN" 실패 : "ERR"
//각 인자의 구분자 ";"
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//  1. 초기화 되지 않은 노드에 접속 시 Initialized = false 전송 (json)																				//
//  2. 안드로이드는 false 수신 시 설정 값을 json으로 전송 이 때 json Oper 항목에 창문 수행 명령이 있을경우 바로 수행								//
//  3. 수행명령이 존재하지 않을 경우 값을 파일로 쓰고 안드로이드로 "OK" 를 json 으로 전송함															//
//																																					//
// 1-1. 초기화 된 노드에 접속시 Initialized = true 전송 (json)																						//
// 2-1. 안드로이드는 true 수신 시 창문에 validation과 동시에 명령을 수행시키기 위해 창문으로 수행 명령 (JSON) Oper항목에 명령을 담아 전송			//
// 2-2. 창문에서 validation 수행 실패 시 "ERR"를 json으로 전송하고 validation 수행 성공 시 Operations 로 넘어가 창문을 조작							//
// 3-1. 수행이 종료되면 "OK"를 json으로 전송																										//
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (win *Window) afterConnected(Android net.Conn, node *Node) {
	//자격증명이 설정되어 있지 않아 자격증명 초기화 시
	if node.PassWord == "" {
		//초기화되어 있지 않음을 Client에 전송
		err := COMM_SENDJSON(&Node{Initialized: false}, Android)
		if err != nil {
			log.Print(err.Error())
		}
		//안드로이드로부터 초기화할 값을 수신함
		recvData, err := COMM_RECVJSON(Android)
		if err != nil {
			log.Print(err.Error())
		}

		//가독성 위해 switch구문 사용
		//지속적인 연결을 위한 operation 함수를 반복하지 않음
		switch win.svrInfo.Initialized {
		//초기화된 경우
		case true:
			//자격증명 수행
			if err := win.svrInfo.HashValidation(recvData.PassWord, MODE_VALIDATION); err != nil {
				_ = COMM_SENDJSON(&Node{Ack: "FAIL"}, Android)
			} else {
				//자격증명 성공시 명령 수행
				win.Available.Lock()
				result := win.Operation(recvData, node)
				_ = COMM_SENDJSON(&result, Android)
				win.Available.Unlock()
			}
		//초기화 되지 않은 경우
		case false:
			//비밀번호를 설정하기 이전 창문 확인을 위해 명령이 수신될 경우
			if recvData.PassWord == "" {
				result := win.Operation(recvData, node)
				_ = COMM_SENDJSON(&result, Android)
			} else {
				//수신된 비밀번호를 설정하되 다중 입력이 들어올 경우 race condition이 발생하므로
				// 1. 가장 먼저 자료를 송신한 app
				// 		+ 파일 설정의 lock을 획득 및 램에 적재된 node객체의 initialized = true로 설정 다른 app의 접근을 제한
				//		+ 송신한 객체에 Oper 자료가 존재하면 해당 명령을 창문에 수행함
				// 2. 이후 늦게 자료를 송신한 app
				// 		+ 파일 설정의 lock을 획득하기 이전 node객체의 initialized = true로 인해 FAIL 수신
				//		+ app입장에서는 연결이 종료되고 다시 창문에 접근해야함(자격증명)
				if win.svrInfo.Initialized {
					win.PErr.Println("Already initialized")
					_ = COMM_SENDJSON(&Node{Ack: "FAIL"}, Android)
					return
				}
				win.FAvailable.Lock()
				win.svrInfo.DATA_INITIALIZER(recvData, true)
				if err := win.svrInfo.FILE_FLUSH(); err != nil {
					win.PErr.Println(err)
				}
				if recvData.Oper != "" {
					result := win.Operation(recvData, node)
					_ = COMM_SENDJSON(&result, Android)
				}
				win.FAvailable.Unlock()
			}
		}

		//들어온 json에 창문 명령이 존재 할 경우
		//TODO: if 구문 정리 필요
		if recvData.Oper != "" {
			win.Available.Lock()
			result := win.Operation(recvData, node)
			_ = COMM_SENDJSON(&result, Android)

			/* 반복적인 창문 여닫는 명령 수행구문 ****deprecated**** TODO: 지속적인 명령을 수신하여 수행할지 검토 필요
			win.Operations(Android, &recvData, node)
			_ = COMM_SENDJSON(&Node{Ack: COMM_SUCCESS}, Android)
			*/

			win.Available.Unlock()
			return
		} else {
			if recvData.PassWord == "" {

			} else {
				//들어온 json에 창문 명령이 존재하지 않을 경우
				node.DATA_INITIALIZER(recvData, false)
				if node.Initialized {
					win.PInfo.Println("Socket server successfully configured (DATA_INITIALIZER)")
				} else {
					win.PInfo.Println("Socket server configuration failed(DATA_INITIALIZER)")
					return
				}

				//입력된 값을 이용하여 초기화, racecondition으로 락킹 적용, 파일 출력
				fileLock := new(sync.Mutex)
				fileLock.Lock()
				err = node.FILE_FLUSH()
				if err != nil {
					win.PErr.Println("Socket server failed to flush (FILE_FLUSH)")
					_ = COMM_SENDJSON(&Node{Ack: "SmartWindow failed to write, reboot please"}, Android)
					return
				}
				fileLock.Unlock()
				win.PInfo.Println("Socket server file write succeeded (FILE_FLUSH)")
				_ = COMM_SENDJSON(&Node{Ack: "OK"}, Android)
			}
		}
		//자격증명 필요
	} else {
		//현재 노드상태 전송
		err := COMM_SENDJSON(&Node{Initialized: node.Initialized, Identity: node.Identity}, Android)
		if err != nil {
			fmt.Println(err.Error())
		}
		//Client 데이터 수신
		recvData, err := COMM_RECVJSON(Android)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//수신된 데이터를 기반으로 Auth 실행
		err = node.Authentication(&recvData)
		if err != nil {
			_ = COMM_SENDJSON(&Node{Ack: COMM_ERR}, Android)
			return
		}
		//창문 명령 실행
		win.Available.Lock()
		win.Operations(Android, &recvData, node)
		_ = COMM_SENDJSON(&Node{Ack: COMM_SUCCESS}, Android)
		win.Available.Unlock()
	}
	win.PInfo.Println("Socket server connection terminated with :" + Android.RemoteAddr().String())
	_ = Android.Close()
}

func (win *Window) updateToRelaySVR() {
	win.quitSIGNAL = make(chan os.Signal, 1)
	go func() {
	loop:
		for {
			sig := <-win.quitSIGNAL
			switch sig {
			case syscall.SIGTERM:
				break loop

			default:
				conn, err := net.Dial("TCP", RELAYSVRIPADDR)
				if err != nil {
					win.PErr.Println("connection err with relaySVR: " + err.Error())
				}
				time.Sleep(2 * time.Second)
				_ = COMM_SENDJSON(&Node{Oper: STATE_ONLINE}, conn)
				inNode, err := COMM_RECVJSON(conn)
				if err == nil {
					win.RelayOperation(inNode, win.svrInfo)
				}
			}
		}
	}()
}

func (win *Window) close_updateToRerlaySVR() {
	win.PInfo.Println("RelaySVR communication is now ineffect")
	signal.Notify(win.quitSIGNAL, syscall.SIGTERM)
}

func (win *Window) RelayOperation(reqNode Node, SvrNode *Node) {

}

func (win *Window) SocketOperation(Android net.Conn, SvrNode *Node) {
	for {
		AndroidNode, err := COMM_RECVJSON(Android)
		if err != nil {
			break
		}
		result := win.Operation(AndroidNode, SvrNode)
		_ = COMM_SENDJSON(&result, Android)
	}

}

//창문 명령
func (win *Window) Operation(order Node, svrNode *Node) Node {
	switch order.Oper {
	case OPERATION_OPEN:

		return Node{Ack: COMM_SUCCESS}
	case OPERATION_CLOSE:

		return Node{Ack: COMM_SUCCESS}
	case OPERATION_INFORMATION:
		//TODO : 센서값 모두 파싱
		return Node{}
	case OPERATION_MODEAUTO:
		svrNode.ModeAuto = order.ModeAuto
		return Node{Ack: COMM_SUCCESS}
	case OPERATION_PROXY:
		svrNode.ModeProxy = order.ModeProxy
		return Node{Ack: COMM_SUCCESS}
	default:
		win.PErr.Println("Not available operation argument ")
	}
	return Node{Ack: COMM_FAIL}
}

//창문 명령
func (win *Window) Operations(Android net.Conn, AndroidNode *Node, SvrNode *Node) {
	//TODO: Error 조건이 센서에 있다면 Error 추가하기
	isBreak := false
	var node Node
	node = *AndroidNode
	for {
		switch AndroidNode.Oper {
		case OPERATION_OPEN:
			win.PInfo.Println("Socket server executed command : OPEN")
			COMM_SENDJSON(&Node{Ack: COMM_SUCCESS}, Android)
		case OPERATION_CLOSE:
			win.PInfo.Println("Socet server executed command : CLOSE")
		case OPERATION_INFORMATION:
			// sensorData ;= EXEC_COMMAND("") TODO : 모든 센서 값 출력하는 쉘 절대경로 작성 및 센서별 입력순서 파악
			//splitedData := strings.Split(sensorData,",")
			win.PInfo.Println("Socket server executed command : INFO")
		case OPERATION_MODEAUTO:
			if AndroidNode.ModeAuto {
				SvrNode.ModeAuto = true
				win.PInfo.Println("Socket server executed command : WINDOW_MODE_AUTO=TRUE")
			} else {
				SvrNode.ModeAuto = false
				win.PInfo.Println("Socket server executed command : WINDOW_MODE_AUTO=FALSE")
			}
		case OPERATION_PROXY:
			SvrNode.ModeProxy = AndroidNode.ModeProxy
			COMM_SENDJSON(&Node{Ack: COMM_SUCCESS}, Android)
		case COMM_SUCCESS: //TODO : ANDROID OK로 수정함
			isBreak = true
			break

		default:
			win.PErr.Println("Socket server received not compatible command (OPER)")
		}
		if isBreak {
			break
		} else {
			node, _ = COMM_RECVJSON(Android)
			AndroidNode = &node
		}
	}
}

//TCP 메시지 전송
func COMM_SENDMSG(msg string, Android net.Conn) error {
	_, err := Android.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("SocketSVR failed to send message")
	}
	return nil
}

//TCP 메시지 수신
func COMM_RECVMSG(android net.Conn) (string, error) {
	inStream := make([]byte, 4096)
	var err error
	_, err = android.Read(inStream)
	if len(inStream) == 0 {
		return "", fmt.Errorf("COMM_SVR : SocketSVR received empty message")
	}
	if err == io.EOF {
		return "EOF", nil
	}

	n := bytes.IndexByte(inStream, 0)
	return string(inStream[:n]), nil
}

//JSON파일 전송
func COMM_SENDJSON(windowData *Node, android net.Conn) error {
	marshalledData, err := json.Marshal(windowData)
	if err != nil {
		return fmt.Errorf("COMM_SVR : SocketSVR Marshalled failed")
	}
	_, _ = android.Write(marshalledData)
	return nil
}

//JSON파일 수신
func COMM_RECVJSON(android net.Conn) (res Node, err error) {
	inStream := make([]byte, 4096)
	tmp := Node{}
	n, err := android.Read(inStream)
	if err != nil {
		return res, fmt.Errorf("COMM_SVR : SocketSVR failed to receive message")
	}
	err = json.Unmarshal(inStream[:n], &tmp)
	if err != nil {
		return res, fmt.Errorf("COMM_SVR : SocketSVR failed to Unmarshaling data stream")
	}
	return tmp, nil
}

//OS 명령 실행
func (win *Window) EXEC_COMMAND(comm string) string {
	out, err := exec.Command("/bin/bash", "-c", comm).Output()
	if err != nil {
		win.PInfo.Println("Socket server failed to run command")
	}
	return string(out)
}

//프로그램 시작부
func (win *Window) Start(address string, port string) error {
	//구조체 객체 선언
	win.svrInfo = &Node{}
	win.PErr = log.New(os.Stdout, color.RedString("ERR : "), log.LstdFlags)
	win.PInfo = log.New(os.Stdout, "INFO : ", log.LstdFlags)
	win.Available = new(sync.Mutex)
	win.FAvailable = new(sync.Mutex)

	if err := win.svrInfo.FILE_INITIALIZE(); err != nil {
		win.PErr.Println(err)
	} else {
		win.PInfo.Println("File loaded")
	}
	//서버 리스닝 시작부
	Android, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		win.PErr.Println("failed to open socket")
		return err
	} else {
		win.PInfo.Println("Socket server initialized = " + address + ":" + port)
		win.svrInfo.PrintData()
	}

	defer func() {
		err := Android.Close()
		if err != nil {
			win.PErr.Println("Socket server terminated abnormaly" + Android.Addr().String())
		}
	}()

	for {
		connect, err := Android.Accept()
		if err != nil {
			win.PErr.Println("Socket server failed to connect TCP with :" + connect.RemoteAddr().String())
		} else {
			win.PInfo.Println("Socket server successfully TCP connected with :" + connect.RemoteAddr().String())
			//start go routine
			go win.afterConnected(connect, win.svrInfo)
		}
		defer func() {
			err := connect.Close()
			if err != nil {
				win.PErr.Println("Socket server connection terminated abnormaly with client :" + Android.Addr().String())
			}
		}()
	}
}
