package InteractiveSocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"sync"
)

//고정 변수
const (
	//서버 포트
	SVRLISTENINGPORT = "6866"
	//수신 명령옵션
	OPERATION_OPEN     = "OPEN"
	OPERATION_CLOSE    = "CLOSE"
	OPERATION_MODEAUTO = "AUTO"
	ERR_SELECTION      = "ERRSELECT"
	COMM_DISCONNECTED  = "EOF"
	ANDROID_ERR        = "ERR"
)

//TCP 연결 수립된 스레드
func afterConnected(Android net.Conn, lock *sync.Mutex, node *NodeData) {
	var androidData string
	var splitedAndroidData []string
	//HELLO가 오지 않을경우 자격증명을 실행하지 않음
	/*if clientMsg := COMM_RECVMSG(Android); clientMsg != "HELLO" {
		fmt.Println("SocketSVR Answerd scanning")
	} else {*/
	//자격증명
	if node.passWord == "" {
		//자격증명 초기화
		COMM_SENDMSG("CONFIG_REQUIRE", Android)
		androidData = COMM_RECVMSG(Android)
		splitedAndroidData = strings.Split(androidData, ";")
		if len(splitedAndroidData) < 2 {
			COMM_SENDMSG("CONFIG_REQUIRE", Android)
			fmt.Println("ERR!! SocketSVR received empty configuration data terminate connection with :" + Android.RemoteAddr().String() + "(might caused search function)")
			_ = Android.Close()
			return
		}
		_ = node.HashValidation(splitedAndroidData[0], MODE_PASSWDCONFIG)
		node.hostName = splitedAndroidData[POS_HOSTNAME]
		fmt.Println("SocketSVR Configuration Succeeded")
		lock.Lock()
		err := node.FILE_FLUSH()
		if err != nil {
			fmt.Println("ERR!! SocketSVR failed to flush")
		}
		lock.Unlock()
		fmt.Println("SocketSVR FILE write Succeeded")

		//창문구동
		Operations(Android)
	} else {
		//자격증명 필요
		COMM_SENDMSG("IDENTIFICATION_REQUIRE:"+node.hostName, Android)
		androidData = COMM_RECVMSG(Android)
		splitedAndroidData = strings.Split(androidData, ";")
		if err := node.HashValidation(splitedAndroidData[POS_PASSWORD], MODE_VALIDATION); err != nil {
			//자격증명 실패
			fmt.Println("ERR!! SocketSVR Client validation failed ")
			COMM_SENDMSG(ANDROID_ERR, Android)
		} else {
			//자격증명 성공
			fmt.Println("SocketSVR Client " + Android.RemoteAddr().String() + " successfully logged in")
			COMM_SENDMSG("LOGEDIN", Android)
			Operations(Android)
		}
	}
	//}

	fmt.Println("SocketSVR Connection terminated with :" + Android.RemoteAddr().String())
	_ = Android.Close()
}

func Operations(Android net.Conn) {
	var operation string

connectionloop:
	for true {
		operation = COMM_RECVMSG(Android)
		switch operation {
		case OPERATION_OPEN:
			fmt.Println("SocketSVR command open execudted")
		case OPERATION_CLOSE:
			fmt.Println("SocketSVR command close executed")
		case OPERATION_MODEAUTO:
			fmt.Println("SocketSVR command auto executed")
			/*
				WindowData, err := COMM_RECVJSON(Android)
				if err != nil {
					fmt.Println(err)
					COMM_SENDMSG(ANDROID_ERR, Android)
				}*/
			//TODO : 수신된 윈도우 설정값 적용

		case COMM_DISCONNECTED:
			break connectionloop

		default:
			COMM_SENDMSG(ERR_SELECTION, Android)
			fmt.Println("ERR!! SocketSVR Received not exist operation")
		}
	}

}

//TCP 메시지 전송
func COMM_SENDMSG(msg string, Android net.Conn) string {

	_, err := Android.Write([]byte(msg))
	if err != nil {
		return "SocketSVR While Send MSG (MSG :" + msg + ")"
	}

	return "netOK"
}

//TCP 메시지 수신
func COMM_RECVMSG(android net.Conn) string {
	inStream := make([]byte, 4096)
	var err error
	for true {
		_, err = android.Read(inStream)
		if len(inStream) != 0 {
			break
		}
	}
	if err == io.EOF {
		return "EOF"
	}

	n := bytes.IndexByte(inStream, 0)
	return string(inStream[:n])
}

func COMM_SENDJSON(windowData *NodeData, android net.Conn) error {
	marshalledData, err := json.Marshal(windowData)
	if err != nil {
		return fmt.Errorf("ERR!! SocketSVR Marshalled failed")
	}
	COMM_SENDMSG(string(marshalledData), android)
	return nil
}

func COMM_RECVJSON(android net.Conn) (*NodeData, error) {
	inStream := make([]byte, 4096)
	var n int
	var err error
	var Node *NodeData

readLoop:
	for true {
		n, err = android.Read(inStream)
		switch err {
		case nil:
			break readLoop

		case io.EOF:
			return nil, fmt.Errorf("ERR!! SocketSVR failed to read JSON")
		}
	}
	err = json.Unmarshal(inStream[:n], Node)
	if err != nil {
		return nil, fmt.Errorf("ERR!! SocketSVR Unmarshal failed")
	}

	return Node, nil
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
	fileInfo := new(NodeData)
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
