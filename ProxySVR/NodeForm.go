package ProxySVR

type NodeForm struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Initialized bool   `json:"Initialized"`
	PassWord    string `json:"PassWord"`
	IPAddr      string `json:"IPAddr"`
	Hostname    string `json:"Hostname"`
	ModeAuto    bool   `json:"ModeAuto"`
	Oper        string `json:"Oper"`
	Temp        int    `json:"Temp"`
	Humidity    int    `json:"Humidity"`
	Gas         int    `json:"Gas"`
	Light       int    `json:"Light"`
}
