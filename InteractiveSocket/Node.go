package InteractiveSocket

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

const (
	FILENAME          = "WindowDATA.txt"
	MODE_PASSWDCONFIG = "PASSWDCONFIG"
	MODE_VALIDATION   = "VALIDATION"
	POS_PASSWORD      = 0
	POS_HOSTNAME      = 1
)

type NodeData struct {
	passWord  string
	hostName  string
	mode_AUTO bool
	temp      int
	humidity  int
	gas       int
	light     int
}

//확인
func (node *NodeData) HashValidation(passwd string, operation string) error {
	hashfunc := md5.New()
	switch operation {
	case MODE_PASSWDCONFIG:
		hashfunc.Write([]byte(passwd))
		node.passWord = hex.EncodeToString(hashfunc.Sum(nil))
	case MODE_VALIDATION:
		hashfunc.Write([]byte(passwd))
		if node.passWord != hex.EncodeToString(hashfunc.Sum(nil)) {
			return fmt.Errorf("ERR!! Password validation failed")
		}
	}
	return nil
}

//파일 확인 및 생성
func (node *NodeData) FILE_INITIALIZE() (bool, error) {
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
		fmt.Println("SocketSVR Found local datafile commencing data initialize")
		content, _ := node.FILE_READ()
		if content == "" {
			return false, fmt.Errorf("SocketSVR found empty data file")
		}
		splitedContent := strings.Split(content, ";")
		if len(splitedContent) == 1 {
			node.passWord = string(splitedContent[POS_PASSWORD])
		} else {
			node.passWord = string(splitedContent[POS_PASSWORD])
			node.hostName = string(splitedContent[POS_HOSTNAME])
		}
		fmt.Println("SocketSVR Data initialize completed")
		return true, nil
	}
	return true, nil

}

//파일 출력
func (node *NodeData) FILE_WRITE(data string) error {
	dataFile, err := os.OpenFile(FILENAME, os.O_RDWR|os.O_TRUNC, os.FileMode(0644))
	defer dataFile.Close()
	if _, err = dataFile.Write([]byte(data)); err != nil {
		return err
	}
	return nil
}

//파일 입력
func (node *NodeData) FILE_READ() (string, error) {
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
func (node *NodeData) FILE_FLUSH() error {
	err := node.FILE_WRITE(node.passWord + ";" + node.hostName)
	return err
}
