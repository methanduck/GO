package InteractiveSocket

import (
	"fmt"
	"net"
)

func afterConnected(Android net.Conn) int {
	flag := -3
	AndroidData := make([]byte, 4096)
	for true {
		AndroidTalk, err := Android.Read(AndroidData);
		if err != nil {
			fmt.Println("SocketSVR Read FAIL")
			flag =1
		}else if string(AndroidData) == "BYE" {
			fmt.Println("SocketSVR Connection Closed (Cause : Client Request)")
			flag = 0;
			break;
		}

		switch string(AndroidData) {
		case "HELLO":
			COMM_SENDMSG("HELLO",Android)

		case "":
			
		}

	}
}

func COMM_SENDMSG(msg string, connector net.Conn) string {

	_, err := connector.Write([]byte(msg))
	if err != nil{
		return "SocketSVR While Send MSG (MSG :"+msg+")"
	}
	
	return "netOK"
}

func main() int {
	Android, err := net.Listen("tcp",":6866")
	if err != nil {
		fmt.Println("SocketSVR Open FAIL")
		return 1
	}
	defer Android.Close();

	for true {
		connect, err := Android.Accept()
		if err != nil{
			fmt.Println("SocketSVR TCP CONN FAIL")
		}
		defer connect.Close()

		go afterConnected(connect)
	}
	fmt.Println()
	return 0;
}
