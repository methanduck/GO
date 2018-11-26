package InteractiveSocket

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"sync"
)

//고정 변수
const (
	FILENAME         = "WindowDATA.txt"
	SVRLISTENINGPORT = "6866"
)

//TCP 연결 수립된 스레드
func afterConnected(Android net.Conn, Node *NodeData, file *os.File, mutex *sync.Mutex) {
	flag := false
	var AndroidData *[]byte
	var length int

	//양방향 통신 시작
	for true {
		AndroidData, length = COMM_RECVMSG(Android, 0)
		if AndroidData == nil {
			return
		}
		switch string((*AndroidData)[:length]) {
		// TODO : 양방향 소켓통신에서 사용될 기능 추가
		case "HELLO":
			if Node == nil {
				fmt.Println("SocketSVR Received HELLO")
				COMM_SENDMSG("SMARTWINDOW Require HostName", Android)
				AndroidData, length = COMM_RECVMSG(Android, 0)
				Node = new(NodeData)
				mutex.Lock()
				//메모리에 할당
				Node.HostName = string((*AndroidData)[:length])
				//디스크에 할당
				FILE_WRITE(Node.HostName, file)
				mutex.Unlock()
				fmt.Println("SocketSVR Configured HOSTNAME : " + Node.HostName + ")")
				COMM_SENDMSG("SocketSVR hostname configuration was successful : "+Node.HostName, Android)
			} else {
				COMM_SENDMSG("ERR!! SocketSVR Rejected HELLO (Cause : Already Configured HOSTNAME)", Android)
				fmt.Println("ERR!! SocketSVR Rejected HELLO (Cause : Already Configured HOSTNAME)")
			}
			break

		case "BYE":
			fmt.Println("ERR!! SocketSVR Connection Closed (Cause : Client Request)")
			COMM_SENDMSG("BYE!", Android)
			flag = true
			break
		}
		if flag {
			fmt.Println("SocketSVR Connection Terminated")
			break
		}
	}
}

func COMM_SENDMSG(msg string, Android net.Conn) string {

	_, err := Android.Write([]byte(msg))
	if err != nil {
		return "SocketSVR While Send MSG (MSG :" + msg + ")"
	}

	return "netOK"
}

func COMM_RECVMSG(Android net.Conn, Count int) (*[]byte, int) {
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

func FILE_CHK() (string, bool, *os.File) {
	var strContent string
	var flag bool
	var file *os.File

	file, err := os.OpenFile(FILENAME, os.O_CREATE|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		fmt.Println("ERR!! SocketSVR Failed to open DATAFILE ")
		flag = false
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	len, err := os.Stat(FILENAME)
	content := make([]byte, len.Size())
	reader.Read(content)
	strContent = string(content)
	if strContent != "" {
		flag = true
	}

	return strContent, flag, file
}

func FILE_WRITE(data string, file *os.File) {
	err := ioutil.WriteFile(file.Name(), []byte(data), os.FileMode(644))
	if err != nil {
		fmt.Println("ERR!! SocketSVR File write failed")
	}
}

func EXEC_COMMAND(comm string) string {
	out, err := exec.Command("/bin/bash", "-c", comm).Output()
	if err != nil {
		fmt.Println("ERR!! SocketSVR Failed to run command")
	}
	return string(out)
}

func Start() int {
	//Load from local file
	Data, result, file := FILE_CHK()
	var Node *NodeData
	lock := new(sync.Mutex)
	//result != file is success of load local file
	if result == true {
		tmpNode := new(NodeData)
		tmpNode.HostName = Data
		Node = tmpNode
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
			go afterConnected(connect, Node, file, lock)
		}
		defer connect.Close()

	}
	return 0
}
