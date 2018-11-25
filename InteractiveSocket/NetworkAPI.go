package InteractiveSocket

import (
	"fmt"
	"net"

)

func afterConnected(Android net.Conn, Node *NodeData)  {
	fmt.Println("afterconnected comes in")
	flag := -3
	var AndroidData *[]byte
	var length int
	for true {
		AndroidData,length = COMM_RECVMSG(Android,0)
		fmt.Println(length)
		if AndroidData == nil {
			return
		}
			switch string((*AndroidData)[:length]) {
			case "HELLO":
				if Node == nil{
					fmt.Println("SocketSVR Received HELLO")
					_ = COMM_SENDMSG("Require HostName",Android)

					AndroidData,length = COMM_RECVMSG(Android,0)
					Node = new(NodeData)
					Node.HostName = string((*AndroidData)[:length])
					fmt.Println("SocketSVR Configured HOSTNAME : "+Node.HostName+")")
				} else {
					fmt.Println("SocketSVR Rejected HELLO (Cause : Already Configured HOSTNAME")
				}
			break

			case "BYE":
				fmt.Println("SocketSVR Connection Closed (Cause : Client Request)")
				flag = 0
				break
			}
		if flag == 0{
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
		fmt.Println("SocketSVR Retrying read MSG ABORTED (Cause : count out)")
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



func Start() int {
	var Node *NodeData

	Android, err := net.Listen("tcp",":6866")
	if err != nil {
		fmt.Println("SocketSVR Open FAIL")
		return 1
	} else {
		fmt.Println("SocketSVR Open Succedded")
	}
	defer Android.Close()

	for  {
		connect, err := Android.Accept()
		if err != nil{
			fmt.Println("SocketSVR TCP CONN FAIL")
		} else {
			fmt.Println("SocketSVR TCP CONN Succeeded")
			go afterConnected(connect, Node)
		}
		defer connect.Close()
	}
	return 0
}
