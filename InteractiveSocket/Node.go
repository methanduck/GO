package InteractiveSocket

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	//수신 명령옵션
	OPERATION_OPEN        = "OPEN"
	OPERATION_CLOSE       = "CLOSE"
	OPERATION_MODEAUTO    = "AUTO"
	OPERATION_INFORMATION = "INFO"
	OPERATION_PROXY       = "REMOTE"
	ERR_SELECTION         = "ERRSELECT"
	COMM_ERR              = "ERR"  // 유효하지 않은 명령이거나 서비스되고 있는 항목이 아닐 경우 ex) 존재하지 않는 동작 명령어 수신 시
	COMM_FAIL             = "FAIL" // 유효한 명령이었으나 정상적으로 처리되지 않은 경우 ex) 로그인 실패
	COMM_SUCCESS          = "OK"   // Transaction 성공 시 전송
	DELIMITER             = ","
	FILENAME              = "WindowDATA.txt"
	MODE_PASSWDCONFIG     = "PASSWDCONFIG"
	MODE_VALIDATION       = "VALIDATION"
	POS_PASSWORD          = 0
	POS_HOSTNAME          = 1
)

type Node struct {
	Which       bool   `json:"which"`       //창문 또는 어플리케이션을 구분 // 창문 : false , 어플리케이션 : true
	Initialized bool   `json:"Initialized"` //TODO: 바로 인증 전송
	PassWord    string `json:"PassWord"`    //창문 비밀번호
	IPAddr      string `json:"IPAddr"`      //TODO: 항목 검토필요
	Identity    string `json:"Identity"`    //중복되는 IP에서도 창문을 구별 할 수 있음
	ModeAuto    bool   `json:"ModeAuto"`    //자동 모드 설정
	ModeProxy   bool   `json:"mode_proxy"`
	Oper        string `json:"Oper"`     // "OPEN", "CLOSE", "CONF"
	Ack         string `json:"Ack"`      // "OK", "COMM_SUCCESS", "TRUE", "FAIL", "FALSE"
	Temp        int    `json:"Temp"`     // temperature
	Humidity    int    `json:"Humidity"` // Humidity
	Gas         int    `json:"Gas"`      // Gas
	Light       int    `json:"Light"`    // Light
}

func (node *Node) Authentication(input *Node) error {
	if node.PassWord != input.PassWord {
		return fmt.Errorf(COMM_ERR)
	}
	return nil
}

//자격증명
func (node *Node) HashValidation(passwd string, operation string) error {
	hashfunc := md5.New()
	switch operation {
	case MODE_PASSWDCONFIG:
		hashfunc.Write([]byte(passwd))
		node.PassWord = hex.EncodeToString(hashfunc.Sum(nil))
	case MODE_VALIDATION:
		hashfunc.Write([]byte(passwd))
		if node.PassWord != hex.EncodeToString(hashfunc.Sum(nil)) {
			return fmt.Errorf("ERR!! Password validation failed")
		}
	}
	return nil
}

//Identity 설정
func (node *Node) SetIdentity() error {
	hashfunc := md5.New()
	addr := node.GetMacAddr()
	hashfunc.Write([]byte(addr))
	node.Identity = hex.EncodeToString(hashfunc.Sum(nil))
	return fmt.Errorf("Failed to set identity")
}

//파일 확인 및 생성
func (node *Node) FILE_INITIALIZE() (bool, error) {
	if _, infoerr := os.Stat(FILENAME); infoerr != nil {
		//로컬 파일이 존재하지 않을 경우
		dataFile, err := os.OpenFile(FILENAME, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, os.FileMode(0644))
		if err != nil {
			return false, err
		}
		defer dataFile.Close()
		fmt.Println("SocketSVR file open succeeded data initialize require!!")
	} else {
		//로컬 파일이 존재 할 경우
		fmt.Println("SocketSVR Found local datafile, commencing data initialize")
		content, _ := node.FILE_READ()
		if content == "" {
			return false, fmt.Errorf("SocketSVR found empty data file, need configuration")
		}
		splitedContent := strings.Split(content, ";")
		if len(splitedContent) == 1 {
			node.PassWord = string(splitedContent[POS_PASSWORD])
		} else {
			node.PassWord = string(splitedContent[POS_PASSWORD])
			node.Identity = string(splitedContent[POS_HOSTNAME])
		}
		fmt.Println("SocketSVR Data initialize completed")
		return true, nil
	}
	return true, nil

}

//파일 출력
func (node *Node) FILE_WRITE(data string) error {
	dataFile, err := os.OpenFile(FILENAME, os.O_RDWR|os.O_TRUNC, os.FileMode(0644))
	defer dataFile.Close()
	if _, err = dataFile.Write([]byte(data)); err != nil {
		return err
	}
	return nil
}

//파일 입력
func (node *Node) FILE_READ() (string, error) {
	dataFile, err := os.OpenFile(FILENAME, os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		return "", err
	}
	defer dataFile.Close()

	fileInfo, _ := os.Stat(FILENAME)
	content := make([]byte, fileInfo.Size())
	dataFile.Read(content)
	return string(content), nil
}

//설정된 값들을 파일로 출력
func (node *Node) FILE_FLUSH() error {
	err := node.FILE_WRITE(node.PassWord + ";" + node.Identity)
	return err
}

//입력된 구조체의 값을 적용
//mode false; 모든 자료 초기화 true ; 자동화 설정 위함
func (node *Node) DATA_INITIALIZER(inputData Node, mode bool) {
	node.Initialized = true
	node.PassWord = inputData.PassWord
	node.Identity = inputData.Identity
	node.ModeAuto = inputData.ModeAuto
	node.Oper = inputData.Oper
	node.Temp = inputData.Temp
	node.Humidity = inputData.Humidity
	node.Gas = inputData.Gas
	node.Light = inputData.Light
}

//Mac addr 수집
func (node *Node) GetMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}
