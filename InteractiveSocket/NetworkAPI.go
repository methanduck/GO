package InteractiveSocket

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"io/ioutil"
	"sync"
)

const (
	FILENAME = "WindowDATA.txt"
	SVRLISTENINGPORT = "6866"
)

func afterConnected(Android net.Conn, Node *NodeData,file *os.File)  {
	fmt.Println("afterconnected comes in")
	flag := -3
	var AndroidData *[]byte
	var length int
	var mutex = new(sync.Mutex)

	for true {
		AndroidData,length = COMM_RECVMSG(Android,0)
		if AndroidData == nil {
			return
		}
			switch string((*AndroidData)[:length]) {
				case "HELLO":
					if Node == nil{
						fmt.Println("SocketSVR Received HELLO")
						COMM_SENDMSG("SMARTWINDOW Require HostName",Android)
						AndroidData,length = COMM_RECVMSG(Android,0)
						Node = new(NodeData)
						//setHOSTNAME on MainMemory
						Node.HostName = string((*AndroidData)[:length])
						//setHOSTNAME on Drive
						mutex.Lock()
						FILE_WRITE(Node.HostName,file)
						mutex.Unlock()
						fmt.Println("SocketSVR Configured HOSTNAME : "+Node.HostName+")")
						COMM_SENDMSG("SocketSVR hostname configuration was successful : "+Node.HostName,Android)
					} else {
						COMM_SENDMSG("ERR!! SocketSVR Rejected HELLO (Cause : Already Configured HOSTNAME)",Android)
						fmt.Println("ERR!! SocketSVR Rejected HELLO (Cause : Already Configured HOSTNAME)")
					}
				break

				case "BYE":
					fmt.Println("ERR!! SocketSVR Connection Closed (Cause : Client Request)")
					COMM_SENDMSG("BYE!",Android)
					flag = 0
				break
			}
		if flag == 0{
			fmt.Println("SocketSVR Connection Terminated")
			break
		}
	}
}

func COMM_SENDMSG(msg string, Android net.Conn) string {

	_, err := Android.Write([]byte(msg))
	if err != nil{
		return "SocketSVR While Send MSG (MSG :"+msg+")"
	}
	
	return "netOK"
}

func COMM_RECVMSG(Android net.Conn, Count int) (*[]byte,int) {
	msg := make([]byte,4096)
	count := Count

	if count > 4{
		fmt.Println("SocketSVR Retrying read MSG ABORTED (Cause : COUNT OUT)")
		return nil,-1
	}

	len,err := Android.Read(msg)
	if err != nil {
		var recover *[]byte
		fmt.Println("SocketSVR Msg receive FAIL");
		fmt.Println("SocketSVR Retrying read MSG")
		recover,len = COMM_RECVMSG(Android,count)
		if recover != nil {
			return recover,len
		}
	}
	return &msg,len
}

func FILE_CHK() (string,bool,*os.File) {
	var strContent string
	var flag bool
	var file *os.File

	file, err := os.OpenFile(FILENAME,os.O_CREATE|os.O_RDWR,os.FileMode(0644))
	if err != nil{
		fmt.Println("ERR!! SocketSVR Failed to open DATAFILE ")
		flag = false
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	len, err := os.Stat(FILENAME)
	content := make([]byte,len.Size())
	reader.Read(content)
	strContent = string(content)
	if strContent != "" {
		flag = true
	}

	return strContent,flag,file
}

func FILE_WRITE(data string,file *os.File) {
	err := ioutil.WriteFile(file.Name(),[]byte(data),os.FileMode(644))
	if err != nil {
		fmt.Println("ERR!! SocketSVR File write failed")
	}
}



func Start() int {
	//Load from local file
	Data,result,file := FILE_CHK()
	var Node *NodeData
	//result != file is success of load local file
	if result == true {
			tmpNode := new(NodeData)
			tmpNode.HostName = Data
			Node = tmpNode
	}
	//Start Socket Server
	Android, err := net.Listen("tcp",":"+SVRLISTENINGPORT)
	if err != nil {
		fmt.Println("ERR!! SocketSVR Open FAIL")
		return 1
	} else {
		fmt.Println("SocketSVR Open Succedded")
	}
	defer Android.Close()

	for  {
		connect, err := Android.Accept()
		if err != nil{
			fmt.Println("ERR!! SocketSVR TCP CONN FAIL")
		} else {
			fmt.Println("SocketSVR TCP CONN Succeeded")
			//start go routine
			go afterConnected(connect, Node, file)
		}
		defer connect.Close()

	}
	return 0
}
