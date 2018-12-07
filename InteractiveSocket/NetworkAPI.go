package InteractiveSocket

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	//수신 명령옵션
	OPERATION_OPEN        = "OPEN"
	OPERATION_CLOSE       = "CLOSE"
	OPERATION_MODEAUTO    = "AUTO"
	OPERATION_INFORMATION = "INFO"
	ERR_SELECTION         = "ERRSELECT"
	COMM_DISCONNECTED     = "EOF"
	ANDROID_ERR           = "ERR" // TODO : err는 메시지 형태로 전송
	ANDROID_OK            = "OK"
	DELIMITER             = ","
)

//VALIDATION 성공 : "LOGEDIN" 실패 : "ERR"
//각 인자의 구분자 ";"
/////////////////////////////////////////////////////////////////////////
//  1. 초기화 되지 않은 노드에 접속 시 Initialized = false 전송 (json)
//  2. 안드로이드는 false 수신 시 설정 값을 json으로 전송 이 때 json Oper 항목에 창문 수행 명령이 있을경우 바로 수행 TODO : 지속적인 통신 연결 위해 소켓유지 필요
//  3. 수행명령이 존재하지 않을 경우 값을 파일로 쓰고 안드로이드로 "OK" 를 json 으로 전송함
//
// 1-1. 초기화 된 노드에 접속시 Initialized = true 전송 (json)
// 2-1. 안드로이드는 true 수신 시 창문에 validation과 동시에 명령을 수행시키기 위해 창문으로 수행 명령 (JSON) Oper항목에 명령을 담아 전송
// 2-2. 창문에서 validation 수행 실패 시 "ERR"를 json으로 전송하고 validation 수행 성공 시 Operations 로 넘어가 창문을 조작
// 3-1. 수행이 종료되면 "OK"를 json으로 전송
//////////////////////////////////////////////////////////////////////////
func afterConnected(Android net.Conn, lock *sync.Mutex, node *Node) {
	//자격증명이 설정되어 있지 않아 자격증명 초기화 시
	if node.PassWord == "" {
		//초기화 되어있음을 전송
		err := COMM_SENDJSON(&Node{Initialized: false}, Android)
		if err != nil {
			fmt.Println(err.Error())
		}
		//안드로이드로부터 초기화할 값을 수신함
		recvData, err := COMM_RECVJSON(Android)
		if err != nil {
			fmt.Println(err.Error())
		} else if recvData.Oper != "" {
			//들어온 json에 창문 명령이 존재 할 경우
			//TODO : 지속적으로 통신을 유지하며 창문 개폐
			lock.Lock()
			SvrAck, err := Operations(Android, &recvData, node)
			if err != nil {
				fmt.Println(err.Error())
			}
			lock.Unlock()
			_ = COMM_SENDJSON(&Node{Ack: SvrAck.Ack}, Android)
			return
		} else {
			if recvData.PassWord == "" {

			} else {
				//들어온 json에 창문 명령이 존재하지 않을 경우
				node.DATA_FLUSH(recvData, false)
				if node.Initialized {
					fmt.Println("SocketSVR Configuration Succeeded")
				} else {
					fmt.Println("ERR!! SocketSVR failed to init")
					return
				}

				//입력된 값을 이용하여 초기화, racecondition으로 락킹 적용, 파일 출력
				fileLock := new(sync.Mutex)
				fileLock.Lock()
				err = node.FILE_FLUSH()
				if err != nil {
					fmt.Println("ERR!! SocketSVR failed to flush")
					_ = COMM_SENDJSON(&Node{Ack: "SmartWindow failed to write, reboot please"}, Android)
					return
				}
				fileLock.Unlock()
				fmt.Println("SocketSVR FILE write Succeeded")
				_ = COMM_SENDJSON(&Node{Ack: "OK"}, Android)
			}

		}
	} else {
		//자격증명 필요
		err := COMM_SENDJSON(&Node{Initialized: node.Initialized, Hostname: node.Hostname}, Android)
		if err != nil {
			fmt.Println(err.Error())
		}
		recvData, err := COMM_RECVJSON(Android)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = node.Authentication(&recvData)
		if err != nil {
			_ = COMM_SENDJSON(&Node{Ack: ANDROID_ERR}, Android)
			return
		}
		lock.Lock()
		SvrAck, err := Operations(Android, &recvData, node)
		if err != nil {
			fmt.Println(err.Error())
		}
		lock.Unlock()
		_ = COMM_SENDJSON(SvrAck, Android)
	}
	fmt.Println("SocketSVR Connection terminated with :" + Android.RemoteAddr().String())
	_ = Android.Close()
}

//창문 구동
func Operations(Android net.Conn, AndroidNode *Node, SvrNode *Node) (*Node, error) {
	switch AndroidNode.Oper {
	case OPERATION_OPEN:
		fmt.Println("SocketSVR command open execudted")
		return &Node{Ack: "OK"}, nil
	case OPERATION_CLOSE:
		fmt.Println("SocketSVR command close executed")
		return &Node{Ack: "OK"}, nil
	case OPERATION_INFORMATION:
		// sensorData ;= EXEC_COMMAND("") TODO : 모든 센서 값 출력하는 쉘 절대경로 작성 및 센서별 입력순서 파악
		//splitedData := strings.Split(sensorData,",")
		fmt.Println("SocketSVR command info executed")
		return &Node{Ack: "OK"}, nil
	case OPERATION_MODEAUTO:
		SvrNode.DATA_FLUSH(*AndroidNode, true)
		fmt.Println("SocketSVR command auto executed")
	case COMM_DISCONNECTED:
		break

	default:
		fmt.Println("SocketSVR Received null data")
	}
	return &Node{Ack: ANDROID_ERR}, fmt.Errorf("SocketSVR failed to execute Operation :" + AndroidNode.Oper)
}

//TCP 메시지 전송
func COMM_SENDMSG(msg string, Android net.Conn) error {
	_, err := Android.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("ERR!! SocketSVR failed to send message")
	}
	return nil
}

//TCP 메시지 수신
func COMM_RECVMSG(android net.Conn) (string, error) {
	inStream := make([]byte, 4096)
	var err error
	_, err = android.Read(inStream)
	if len(inStream) == 0 {
		return "", fmt.Errorf("ERR!! SocketSVR received empty message")
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
		return fmt.Errorf("ERR!! SocketSVR Marshalled failed")
	}
	_, _ = android.Write(marshalledData)
	return nil
}

//JSON파일 수
func COMM_RECVJSON(android net.Conn) (res Node, err error) {
	inStream := make([]byte, 4096)
	tmp := Node{}
	n, err := android.Read(inStream)
	if err != nil {
		return res, fmt.Errorf("ERR!! SocketSVR failed to receive message")
	}
	err = json.Unmarshal(inStream[:n], &tmp)
	if err != nil {
		return res, fmt.Errorf("ERR!! SocketSVR failed to Unmarshaling data stream")
	}
	return tmp, nil
}

//OS 명령 실행
func EXEC_COMMAND(comm string) string {
	out, err := exec.Command("/bin/bash", "-c", comm).Output()
	if err != nil {
		fmt.Println("ERR!! SocketSVR Failed to run command")
	}
	return string(out)
}

//프로그램 시작부
func Start() int {
	//변수선언
	var initerr error
	lock := new(sync.Mutex)
	//구조체 객체 선언
	fileInfo := new(Node)
	_, initerr = fileInfo.FILE_INITIALIZE()
	//빈 파일을 불러오거나 파일을 읽기에 실패했을 경우
	if initerr != nil {
		fmt.Println(initerr)
	}

	//Start Socket Server
	Android, err := net.Listen("tcp", ":"+SVRLISTENINGPORT)
	if err != nil {
		fmt.Println("ERR!! SocketSVR Open FAIL")
		return 1
	} else {
		fmt.Println("SocketSVR Open Succedded")
	}
	defer Android.Close()

	for {
		connect, err := Android.Accept()
		if err != nil {
			fmt.Println("ERR!! SocketSVR TCP CONN FAIL")
		} else {
			fmt.Println("SocketSVR TCP CONN Succeeded : " + connect.RemoteAddr().String())
			//start go routine
			go afterConnected(connect, lock, fileInfo)
		}
		defer connect.Close()
	}
	return 0
}
