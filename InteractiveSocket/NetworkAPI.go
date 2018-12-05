package InteractiveSocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"

	//"strings"
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
	REQUIRE_CONFIG     = "0;null"
	REQUIRE_AUTH       = "1"
)

//VALIDATION 성공 : "LOGEDIN" 실패 : "ERR"
//각 인자의 구분자 ";"
//TCP 연결 수립된 스레드
func afterConnected(Android net.Conn, lock *sync.Mutex, node *Node) {
	var androidData string
	var splitedAndroidData []string
	var err error
	//자격증명이 설정되어 있지 않아 자격증명 초기화 시
	if node.PassWord == "" {
		//자격증명 초기화 (자격증명 설정이 되어 있지 않을 경우)
		_ = COMM_SENDMSG(REQUIRE_CONFIG, Android)
		androidData, err = COMM_RECVMSG(Android)
		if err != nil {
			fmt.Println(err.Error())
			Android.Close()
		}
		splitedAndroidData = strings.Split(androidData, ";")
		//적절한 인자가 대입되지 않았을 경우
		if len(splitedAndroidData) < 2 {
			_ = COMM_SENDMSG("CONFIG_REQUIRE", Android)
			fmt.Println("ERR!! SocketSVR received empty configuration data terminate connection with :" + Android.RemoteAddr().String() + "(might caused search function)")
			_ = Android.Close()
			return
		}
		//적절한 인자가 대입되어 램에 데이터 할당
		_ = node.HashValidation(splitedAndroidData[0], MODE_PASSWDCONFIG)
		node.Hostname = splitedAndroidData[POS_HOSTNAME]
		fmt.Println("SocketSVR Configuration Succeeded")
		//입력된 값을 이용하여 초기화, racecondition으로 락킹 적용
		lock.Lock()
		//설정된 값을 파일로 출력
		err = node.FILE_FLUSH()
		if err != nil {
			fmt.Println("ERR!! SocketSVR failed to flush")
		}
		lock.Unlock()
		fmt.Println("SocketSVR FILE write Succeeded")

		//초기화 과정이므로 별도의 자격증명 없이 명령 구동
		lock.Lock()
		Operations(Android, node)
		lock.Unlock()

	} else {
		//자격증명이 설정되어 있어 자격증명 시작
		_ = COMM_SENDMSG(REQUIRE_AUTH+node.Hostname, Android)
		androidData, err := COMM_RECVMSG(Android)
		if err != nil {
			fmt.Println(err.Error())
			_ = Android.Close()
		}
		splitedAndroidData = strings.Split(androidData, ";")
		//적절한 인자가 대입되지 않았을 경우
		if len(splitedAndroidData) < 2 {
			_ = COMM_SENDMSG("CONFIG_REQUIRE", Android)
			fmt.Println("ERR!! SocketSVR received empty configuration data terminate connection with :" + Android.RemoteAddr().String() + "(might caused search function)")
			_ = Android.Close()
			return
		}
		//적절한 인자가 대입되어 validation 과정 진행
		if err := node.HashValidation(splitedAndroidData[POS_PASSWORD], MODE_VALIDATION); err != nil {
			//자격증명 실패
			fmt.Println("ERR!! SocketSVR Client validation failed ")
			_ = COMM_SENDMSG(ANDROID_ERR, Android)
		} else {
			//자격증명 성공
			fmt.Println("SocketSVR Client " + Android.RemoteAddr().String() + " successfully logged in")
			_ = COMM_SENDMSG("LOGEDIN", Android)
			lock.Lock()
			Operations(Android, node)
			lock.Unlock()
		}
	}
	fmt.Println("SocketSVR Connection terminated with :" + Android.RemoteAddr().String())
	_ = Android.Close()
}

//창문 구동
func Operations(Android net.Conn, node *Node) {
	var operation string
	var err error
	var ConfigData Node
connectionloop:
	for true {
		operation, err = COMM_RECVMSG(Android)
		if err != nil {
			fmt.Println(err.Error() + "(Cause : at Operations)")
		}
		switch operation {
		case OPERATION_OPEN:
			fmt.Println("SocketSVR command open execudted")
		case OPERATION_CLOSE:
			fmt.Println("SocketSVR command close executed")
		case OPERATION_MODEAUTO:
			ConfigData, err = COMM_RECVJSON(Android)
			if err != nil {
				fmt.Println(err.Error())
				break connectionloop
			}
			node.DATA_FLUSH(ConfigData)
			fmt.Println("SocketSVR command auto executed")
		case COMM_DISCONNECTED:
			break connectionloop

		default:
			COMM_SENDMSG(ERR_SELECTION, Android)
			fmt.Println("ERR!! SocketSVR Received not exist operation")
		}
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

func COMM_SENDJSON(windowData *Node, android net.Conn) error {
	marshalledData, err := json.Marshal(windowData)
	if err != nil {
		return fmt.Errorf("ERR!! SocketSVR Marshalled failed")
	}
	android.Write(marshalledData)
	return nil
}

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
