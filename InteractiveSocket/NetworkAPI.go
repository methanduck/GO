package InteractiveSocket

import (
	"fmt"
	"net"
	"strings"
	"os"
	"github.com/methanduck/GO/InteractiveSocket"
)

func afterConnected(Android net.Conn, Node *NodeData,file *os.File)  {
	fmt.Println("afterconnected comes in")
	flag := -3
	var AndroidData *[]byte
	var length int
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
					result := FILE_WRITE(Node.HostName,file)
					if result != false{
						fmt.Println("SocketSVR Configured HOSTNAME : "+Node.HostName+")")
					} else {
						fmt.Println("ERR!! SocketSVR Configured HOSTNAME : " + Node.HostName + ") (Cause : Couldn't wrote data on the Drive) ")
					}
				} else {
					fmt.Println("ERR!! SocketSVR Rejected HELLO (Cause : Already Configured HOSTNAME)")
				}
			break

			case "BYE":
				fmt.Println("ERR!! SocketSVR Connection Closed (Cause : Client Request)")
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
	info, err := os.Stat("./WindowDATA.txt")
	if os.IsNotExist(err){
		newDATA, err := os.Create("./WindowDATA.txt")
		if err != nil{
			fmt.Println("SocketSVR Create File FAIL")
			return strContent,false,nil
		}
		defer newDATA.Close()

		return strContent,false,newDATA
	} else {
		DATA, err := os.OpenFile("./",os.O_RDWR,0644)
		if err != nil {
			fmt.Println("SocketSVR Read File FAIL")
		}
		defer DATA.Close()

		content := make([]byte,info.Size())
		n,err := DATA.Read(content)
		if err != nil{
			fmt.Println(fmt.Println("SocketSVR Read File FAIL"))
		}
		strContent =string(content[:n])
		return strContent,true,DATA
	}

}

func FILE_WRITE(data string,file *os.File) bool {
	_,err := file.Write([]byte(data))
	if err != nil {
		fmt.Println("ERR!! SocketSVR Couldn't wrote Data on the Drive")
		return false
	}
	return true
}

func Start() int {
	//Load from local file
	Data,result,file := FILE_CHK()
	var Node *NodeData
	//result != file is success of load local file
	if result != false {
			splited := strings.Split(Data,";")
			tmpNode := new(NodeData)
			tmpNode.HostName = splited[0]
			tmpNode.AndroidIP = splited[1]
			Node = tmpNode
	}
	//Start Socket Server
	Android, err := net.Listen("tcp",":6866")
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
