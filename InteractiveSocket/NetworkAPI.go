package InteractiveSocket

import (
	"bytes"
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
	OPEN              = "OPEN"
	CLOSE             = "CLOSE"
	ERR_SELECTION     = "ERRSELECT"
	COMM_DISCONNECTED = "EOF"
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
		if len(splitedAndroidData) < 2 {
			fmt.Println("ERR!! SocketSVR received empty configuration data terminate connection with :" + Android.RemoteAddr().String() + "(Cause : passwd configuration Function) critical errstate")
			_ = Android.Close()
			return
		}
		_ = node.HashValidation(splitedAndroidData[0], MODE_PASSWDCONFIG)
		node.HostName = splitedAndroidData[POS_HOSTNAME]
		fmt.Println("SocketSVR Configuration Succeeded")
		lock.Lock()
		err := node.FILE_FLUSH()
		if err != nil {
			fmt.Println("ERR!! SocketSVR failed to flush")
		}
		lock.Unlock()
		fmt.Println("SocketSVR FILE write Succeeded")
		//창문구동패

		Operations(Android)
	} else {
		//자격증명 필요
		COMM_SENDMSG("IDENTIFICATION_REQUIRE", Android)
		androidData = COMM_RECVMSG(Android)
		splitedAndroidData = strings.Split(androidData, ";")
		if err := node.HashValidation(splitedAndroidData[POS_PASSWORD], MODE_VALIDATION); err != nil {
			//자격증명 실
			fmt.Println("ERR!! SocketSVR Client validation failed ")
		} else {
			//자격증명 성공
			fmt.Println("SocketSVR Client " + Android.RemoteAddr().String() + " successfully logged in")
			Operations(Android)
		}
	}
	fmt.Println("SocketSVR Connection terminated with :" + Android.RemoteAddr().String())
	_ = Android.Close()
}

func Operations(Android net.Conn) {
	var operation string

connectionloop:
	for true {
		operation = COMM_RECVMSG(Android)
		switch operation {
		case OPEN:

		case CLOSE:

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
			fmt.Println("SocketSVR TCP CONN Succeeded : " + connect.RemoteAddr().String())
			//start go routine
			go afterConnected(connect, lock, nodeInfo)
		}
		defer connect.Close()
	}
	return 0
}
