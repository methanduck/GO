package InteractiveSocket

import (
	"fmt"
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
	OPEN          = "OPEN"
	CLOSE         = "CLOSE"
	ERR_SELECTION = "ERRSELECT"
)

var (
	nodeInfo *NodeData
)

//TCP 연결 수립된 스레드
func afterConnected(Android net.Conn, lock *sync.Mutex, node *NodeData) {
	var androidData string
	var splitedAndroidData []string

	//자격증명
	if node.passWord == "" {
		//자격증명 초기화
		COMM_SENDMSG("CONFIG_REQUIRE", Android)
		androidData = COMM_RECVMSG(Android)
		splitedAndroidData = strings.Split(androidData, ";")
		node.passWord = splitedAndroidData[POS_PASSWORD]
		node.HostName = splitedAndroidData[POS_HOSTNAME]
		lock.Lock()
		node.FILE_FLUSH()
		lock.Unlock()
		//창문구동
		Operations(Android)
	} else {
		//자격증명 필요
		androidData = COMM_RECVMSG(Android)
		splitedAndroidData = strings.Split(androidData, ";")
		if err := node.HashValidation(splitedAndroidData[POS_PASSWORD], MODE_VALIDATION); err != nil {
			//자격증명 실
			fmt.Println("ERR!! SocketSVR Client validation failed ")
		} else {
			//자격증명 성공패
			fmt.Println("SocketSVR Client " + Android.RemoteAddr().String() + " successfully logged in")
			Operations(Android)
		}
	}
	fmt.Println("SocketSVR Connection terminated with :" + Android.RemoteAddr().String())
}

func Operations(Android net.Conn) {
	var operation string
	for true {
		operation = COMM_RECVMSG(Android)
		switch operation {
		case OPEN:

		case CLOSE:

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

func COMM_RECVMSG(android net.Conn) string {
	inStream := make([]byte, 4096)
	for true {
		android.Read(inStream)
		if len(inStream) != 0 {
			break
		}
	}
	return string(inStream)
}

/*
//TCP 메시지 수신 deprecated
func receive(Android net.Conn, Count int) (*[]byte, int) {
	msg := make([]byte, 4096)
	count := Count

	if count > 4 {
		fmt.Println("SocketSVR Retrying read MSG ABORTED (Cause : COUNT OUT)")
		return nil, -1
	}

	len, err := Android.Read(msg)
	if err != nil {
		var recover *[]byte
		fmt.Println("SocketSVR Msg receive FAIL")
		fmt.Println("SocketSVR Retrying read MSG")
		recover, len = COMM_RECVMSG(Android, count)
		if recover != nil {
			return recover, len
		}
	}
	return &msg, len
}
*/

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
	lock := new(sync.Mutex)
	nodeInfo = new(NodeData)
	initerr := nodeInfo.FILE_INITIALIZE()

	if initerr != nil {
		fmt.Println("ERR!! SocketSVR failed to initialize from local file")
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
			fmt.Println("SocketSVR TCP CONN Succeeded")
			//start go routine
			go afterConnected(connect, lock, nodeInfo)
		}
		defer connect.Close()

	}
	return 0
}
