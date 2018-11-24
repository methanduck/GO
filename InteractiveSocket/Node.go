package InteractiveSocket

import "strconv"

type NodeData struct {
	AndroidIP string
	HostName string
	Temp int
	Humidity int
	Gas int
	Light int
}

func(n *NodeData) PrintStatus() string {
	return n.AndroidIP+n.HostName+strconv.Itoa(n.Temp)+strconv.Itoa(n.Humidity)+strconv.Itoa(n.Gas)+strconv.Itoa(n.Light)
}
