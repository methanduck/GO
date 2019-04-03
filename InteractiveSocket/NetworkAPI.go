package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	//"hash/adler32"
	"io"
	"net"
	"os/exec"
	//"strings"

	//"strings"
	"sync"
)

//고정 변수
const (
	//서버 포트
	SVRLISTENINGPORT = "6866"
)

type Window struct {
	PInfo     *log.Logger
	PErr      *log.Logger
	svrInfo   *Node
	Available *sync.Mutex
}

//VALIDATION 성공 : "LOGEDIN" 실패 : "ERR"
//각 인자의 구분자 ";"
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//  1. 초기화 되지 않은 노드에 접속 시 Initialized = false 전송 (json)
//  2. 안드로이드는 false 수신 시 설정 값을 json으로 전송 이 때 json Oper 항목에 창문 수행 명령이 있을경우 바로 수행
//  3. 수행명령이 존재하지 않을 경우 값을 파일로 쓰고 안드로이드로 "OK" 를 json 으로 전송함
//
// 1-1. 초기화 된 노드에 접속시 Initialized = true 전송 (json)
// 2-1. 안드로이드는 true 수신 시 창문에 validation과 동시에 명령을 수행시키기 위해 창문으로 수행 명령 (JSON) Oper항목에 명령을 담아 전송
// 2-2. 창문에서 validation 수행 실패 시 "ERR"를 json으로 전송하고 validation 수행 성공 시 Operations 로 넘어가 창문을 조작
// 3-1. 수행이 종료되면 "OK"를 json으로 전송
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//TODO: 해결됌(Operations을 계속 실행함)
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

		//들어온 json에 창문 명령이 존재 할 경우
		//TODO: 창문에 계정 설정을 먼저 할 지 개폐 먼저 할지 결정
		//TODO: if 구문 정리 필요
		if recvData.Oper != "" {
			win.Available.Lock()
			win.Operations(Android, &recvData, node)
			_ = COMM_SENDJSON(&Node{Ack: COMM_SUCCESS}, Android)
			win.Available.Unlock()
			return
		} else {
			if recvData.PassWord == "" {

			} else {
				//들어온 json에 창문 명령이 존재하지 않을 경우
				node.DATA_INITIALIZER(recvData, false)
				if node.Initialized {
					fmt.Println("SocketSVR Configuration Succeeded")
				} else {
					fmt.Println("COMM_SVR : SocketSVR failed to init")
					return
				}

				//입력된 값을 이용하여 초기화, racecondition으로 락킹 적용, 파일 출력
				fileLock := new(sync.Mutex)
				fileLock.Lock()
				err = node.FILE_FLUSH()
				if err != nil {
					fmt.Println("COMM_SVR : SocketSVR failed to flush")
					_ = COMM_SENDJSON(&Node{Ack: "SmartWindow failed to write, reboot please"}, Android)
					return
				}
				fileLock.Unlock()
				fmt.Println("SocketSVR FILE write Succeeded")
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
	fmt.Println("SocketSVR Connection terminated with :" + Android.RemoteAddr().String())
	_ = Android.Close()
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
			fmt.Println("COMM_SVR : INFO command open execudted")
			COMM_SENDJSON(&Node{Ack: COMM_SUCCESS}, Android)
		case OPERATION_CLOSE:
			fmt.Println("COMM_SVR : INFO command close executed")
		case OPERATION_INFORMATION:
			// sensorData ;= EXEC_COMMAND("") TODO : 모든 센서 값 출력하는 쉘 절대경로 작성 및 센서별 입력순서 파악
			//splitedData := strings.Split(sensorData,",")
			fmt.Println("COMM_SVR : INFO command info executed")
		case OPERATION_MODEAUTO:
			if AndroidNode.ModeAuto {
				SvrNode.ModeAuto = true
				fmt.Print("COMM_SVR : INFO WINDOW_MODE_AUTO=TRUE")
			} else {
				SvrNode.ModeAuto = false
				fmt.Print("COMM_SVR : INFO WINDOW_MODE_AUTO=FALSE")
			}
		case OPERATION_PROXY:
			SvrNode.ModeProxy = AndroidNode.ModeProxy
			COMM_SENDJSON(&Node{Ack: COMM_SUCCESS}, Android)
		case COMM_SUCCESS: //TODO : ANDROID OK로 수정함
			isBreak = true
			break

		default:
			log.Print("COMM_SVR : ERR! Received not compatible command / func Operations")
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
		return fmt.Errorf("COMM_SVR : SocketSVR failed to send message")
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
func EXEC_COMMAND(comm string) string {
	out, err := exec.Command("/bin/bash", "-c", comm).Output()
	if err != nil {
		fmt.Println("COMM_SVR : SocketSVR Failed to run command")
	}
	return string(out)
}

//프로그램 시작부
func (win *Window) Start() int {
	//변수선언
	var initerr error
	win.Available = new(sync.Mutex)
	//구조체 객체 선언
	win.svrInfo = new(Node)
	_, initerr = win.svrInfo.FILE_INITIALIZE()
	//빈 파일을 불러오거나 파일을 읽기에 실패했을 경우
	if initerr != nil {
		fmt.Println(initerr)
	}

	//Start Socket Server
	Android, err := net.Listen("tcp", ":"+SVRLISTENINGPORT)
	if err != nil {
		log.Print("COMM_SVR : ERR Socket Open FAIL")
		return 1
	} else {
		log.Print("COMM_SVR : INFO Socket Open Succedded")
	}
	defer func() {
		err := Android.Close()
		if err != nil {
			log.Print("COMM_SVR: ERR Socket server terminated abnormaly " + Android.Addr().String())
		}
	}()

	for {
		connect, err := Android.Accept()
		if err != nil {
			fmt.Println("COMM_SVR : SocketSVR TCP CONN FAIL")
		} else {
			fmt.Println("SocketSVR TCP CONN Succeeded : " + connect.RemoteAddr().String())
			//start go routine
			go win.afterConnected(connect, win.svrInfo)
		}
		defer func() {
			err := connect.Close()
			if err != nil {
				log.Print("COMM_SVR: ERR! client connection terminated abnormaly " + Android.Addr().String())
			}
		}()
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	localServer := Window{}
	localServer.Start()
}
