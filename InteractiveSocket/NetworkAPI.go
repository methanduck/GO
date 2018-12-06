package InteractiveSocket

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	ANDROID_ERR           = "ERR"
	ANDROID_LOGIN         = "LOGEDIN"
)

//VALIDATION 성공 : "LOGEDIN" 실패 : "ERR"
//각 인자의 구분자 ";"
//TCP 연결 수립된 스레드
func afterConnected(Android net.Conn, lock *sync.Mutex, node *Node) {
	//자격증명이 설정되어 있지 않아 자격증명 초기화 시
	if node.passWord == "" {
		err := COMM_SENDJSON(&Node{Initialized: false}, Android)
		if err != nil {
			fmt.Println(err.Error())
		}
		recvData, err := COMM_RECVJSON(Android)
		if err != nil {
			fmt.Println(err.Error())
		} else if recvData.Oper != "" {
			lock.Lock()
			Operations(Android, &recvData, node)
			fmt.Println("SocketSVR testing window")
			lock.Unlock()
		} else {

			node.DATA_FLUSH(recvData, false)
			if node.Initialized {
				fmt.Println("SocketSVR Configuration Succeeded")
			} else {
				fmt.Println("ERR!! SocketSVR failed to init")
				return
			}

			//입력된 값을 이용하여 초기화, racecondition으로 락킹 적용
			fileLock := new(sync.Mutex)
			fileLock.Lock()
			//설정된 값을 파일로 출력
			err = node.FILE_FLUSH()
			if err != nil {
				fmt.Println("ERR!! SocketSVR failed to flush")
			}
			fileLock.Unlock()
			fmt.Println("SocketSVR FILE write Succeeded")

			//초기화 과정이므로 별도의 자격증명 없이 명령 구동
			lock.Lock()
			Operations(Android, &recvData, node)
			lock.Unlock()
		}
	} else {
		err := COMM_SENDJSON(&Node{Initialized: node.Initialized, Hostname: node.Hostname}, Android)
		if err != nil {
			fmt.Println(err.Error())
		}
		recvData, err := COMM_RECVJSON(Android)
		if err != nil {
			fmt.Println(err.Error())
			return
		} else if recvData.Oper != "" {
			lock.Lock()
			Operations(Android, &recvData, node)
			lock.Unlock()
			fmt.Println("SocketSVR testing window")
		} else {
			err = node.HashValidation(recvData.passWord, MODE_VALIDATION)
			if err != nil {
				fmt.Println("ERR!! SocketSVR client failed to login")
				_ = COMM_SENDMSG(ANDROID_ERR, Android)
			} else {
				fmt.Println("SocketSVR client successfully loged in")
				_ = COMM_SENDMSG(ANDROID_LOGIN, Android)
				Operations(Android, &recvData, node)
			}
		}
	}
	fmt.Println("SocketSVR Connection terminated with :" + Android.RemoteAddr().String())
	_ = Android.Close()
}

//창문 구동
func Operations(Android net.Conn, AndroidNode *Node, SvrNode *Node) {
	switch AndroidNode.Oper {
	case OPERATION_OPEN:
		fmt.Println("SocketSVR command open execudted")
	case OPERATION_CLOSE:
		fmt.Println("SocketSVR command close executed")
	case OPERATION_INFORMATION:
		// sensorData ;= EXEC_COMMAND("") TODO : 모든 센서 값 출력하는 쉘 절대경로 작성 및 센서별 입력순서 파악
		//splitedData := strings.Split(sensorData,",")
		err := COMM_SENDJSON(&Node{}, Android)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("SocketSVR command info executed")

	case OPERATION_MODEAUTO:
		SvrNode.DATA_FLUSH(*AndroidNode, true)
		fmt.Println("SocketSVR command auto executed")
	case COMM_DISCONNECTED:
		return
	default:
		COMM_SENDMSG(ERR_SELECTION, Android)
		fmt.Println("ERR!! SocketSVR Received not exist operation")
	}

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
	android.Write(marshalledData)
	return nil
}

//JSON파일 수
func COMM_RECVJSON(android net.Conn) (res Node, err error) {
	inStream := make([]byte, 4096)
	n, err := android.Read(inStream)
	if err != nil {
		return res, fmt.Errorf("ERR!! SocketSVR failed to receive message")
	}
	err = json.Unmarshal(inStream[:n], res)
	if err != nil {
		return res, fmt.Errorf("ERR!! SocketSVR failed to Unmarshaling data stream")
	}
	return res, nil
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
