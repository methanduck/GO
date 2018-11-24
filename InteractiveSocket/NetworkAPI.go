package InteractiveSocket

import (
	"fmt"
	"net"

)

func afterConnected(Android net.Conn, Node *NodeData)  {
	flag := -3
	var AndroidData *[]byte

	for true {
		AndroidData = COMM_RECVMSG(Android,0)

			switch string(*AndroidData) {
			case "HELLO":
				if Node == nil{
					COMM_SENDMSG("Require HostName",Android)
					AndroidData = COMM_RECVMSG(Android,0)
					Node = new(NodeData)
					Node.HostName = string(*AndroidData)
				}

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

func COMM_RECVMSG(Android net.Conn, Count int) (*[]byte) {
	var msg []byte
	count := Count

	if count > 4{
		fmt.Println("SocketSVR Retrying read MSG ABORTED (Cause : count out)")
		return nil
	}
	_,err := Android.Read(msg)
	if err != nil {
		var recover *[]byte
		fmt.Println("SocketSVR Msg receive FAIL");
		fmt.Println("SocketSVR Retrying read MSG")
		recover = COMM_RECVMSG(Android,count)
		if recover != nil {
			return recover
		}
	}
	return &msg
}



func main() int {
	var Node *NodeData

	Android, err := net.Listen("tcp",":6866")
	if err != nil {
		fmt.Println("SocketSVR Open FAIL")
		return 1
	}
	defer Android.Close()

	for true {
		connect, err := Android.Accept()
		if err != nil{
			fmt.Println("SocketSVR TCP CONN FAIL")
		}
		defer connect.Close()

		go afterConnected(connect, Node)
	}


	return 0
}
