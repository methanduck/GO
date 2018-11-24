package InteractiveSocket

import "strconv"

type NodeData struct {
	AndroidIP string;
	HostName string;
	Temp int;
	Humidity int;
	Gas int;
	Light int;
}

func NewNodeData(hostname string,ip string, temp int, humidity int, gas int, light int) *NodeData {
	NewNode := NodeData{HostName:hostname,AndroidIP:ip,Temp:temp,Humidity:humidity,Gas:gas,Light:light};
	return &NewNode;
}

func(n *NodeData) PrintStatus() string {
	return n.AndroidIP+n.HostName+strconv.Itoa(n.Temp)+strconv.Itoa(n.Humidity)+strconv.Itoa(n.Gas)+strconv.Itoa(n.Light);
}