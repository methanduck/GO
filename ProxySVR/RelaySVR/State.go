package RelaySVR

import (
	"encoding/json"
	"fmt"
	"github.com/GO/InteractiveSocket"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"log"
	"time"
)

const (
	ONLINE       = true
	OFFLINE      = false
	NOTAVAILABLE = "N/A"
	ERRPROCESSING = "PSERR"
	BUCKET_NODE  = "NODE"
)

type dbData struct {
	database *bolt.DB
	pInfo    *log.Logger
	PErr     *log.Logger
}

type NodeState struct {
	Identity        string
	IsOnline        bool
	IsRequireConn   bool
	ApplicationData InteractiveSocket.Node
	Locking         int
}
//기본적으로 온라인인 상태로 리턴
func NodeStateMaker(data InteractiveSocket.Node,state NodeState) NodeState {
	return NodeState{ApplicationData:data,Identity:data.Identity,IsOnline:ONLINE,IsRequireConn:state.IsRequireConn,Locking:0}
}
//Starting BoltDB
//서버초기화에 실패 시 서버가 비정상 종료됩니다.
//Bucket : State, Node
func (db dbData) Startbolt(pinfo *log.Logger, perr *log.Logger) *bolt.DB {
	db.pInfo = pinfo
	db.PErr = perr
	boltdb, err := bolt.Open("SmartWindow.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Panic("Failed to open bolt database")
	}
	db.database = boltdb
	defer func() {
		if err := boltdb.Close(); err != nil {
			perr.Panic("bolt database abnormally terminated")
		}
	}()

	pinfo.Println("BOLT : create new bucket")
	//Node state bucket creation
	if err := boltdb.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("Node")); err != nil {
			return fmt.Errorf("bucket creation failed")
		}
		pinfo.Println("BOLT : bucket creation secceded")
		return nil
	}); err != nil {
		db.PErr.Panic("BOLT : Failed to  Start BoltDB (ERR code :" + err.Error() + ")")
	}
	return boltdb
}
//노드가 존재하는지 확인하고 존재한다면, 온라인 상태를 반환합니다.
//error : 데이터가 존재하지 않습니다, 데이터 처리중 오류가 발생했습니다.
func (db dbData) IsExistAndIsOnline(identity string) (bool,error) {
	var isOnline bool
	tempNodeState := NodeState{}
	if err := db.database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NODE))
		val := bucket.Get([]byte(identity))
		if len(val) == 0 {
			return fmt.Errorf(NOTAVAILABLE)
		}
		if err :=json.Unmarshal(val,tempNodeState); err != nil {
			return fmt.Errorf("BOLT : failed to unmarshal")
		}
		isOnline = tempNodeState.IsOnline
		return nil
	}); err != nil {
		if err.Error() == NOTAVAILABLE {
			return false,errors.New(NOTAVAILABLE)
		}
		return false,errors.New(ERRPROCESSING)
	} else {
		return isOnline,nil
	}
}
//Update the Node "State" as Online
func (db dbData) UpdataOnline(data InteractiveSocket.Node) error {
	var temp_NodeState NodeState
	var val []byte
	isOnline,isExist := db.IsExistAndIsOnline(data.Identity)
	//노드가 데이터베이스에 존재하는지 확인
	//노드가 데이터베이스에 존재 할 때
	if isExist {
		if err := db.database.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(BUCKET_NODE))
			val = bucket.Get([]byte(data.Identity))
			if err := json.Unmarshal(val,&temp_NodeState); err != nil {
				return fmt.Errorf("BOLT : Failed to unmarshal")
			}
			//Modify state
			temp_NodeState.IsOnline = true
			//marshalling
			val,err := json.Marshal(temp_NodeState)
			if err != nil {return fmt.Errorf("BOLT : Failed to marshal")}
			//Update Node
			if err := bucket.Put([]byte(data.Identity),val); err != nil {
				fmt.Println("BOLT : Failed to update NodeState")
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	} else { //Node not exist
		if err := db.database.Update(func(tx *bolt.Tx) error {
			temp_NodeState.ApplicationData = data
			temp_NodeState.
			bucket := tx.Bucket([]byte(BUCKET_NODE))
			val, _ = json.Marshal(temp_NodeState)

		}); err != nil {

		}
	}
}
//Update the Node data
func (db dbData) UpdateNodeData(data *InteractiveSocket.Node)  {

}
//////////////////////////////////////////////////////////////////////////////////////////////////////////
func (db dbData)

func (db dbData) ApplicationNodeUpdate(data InteractiveSocket.Node, locking int) error {
	db.pInfo.Println("BOLT : Commence application request update ")
	state, err := db.GetWindowData(data.Identity)
	if err != nil {
		db.PErr.Println("BOLT : Failed (ERR code :" + err.Error() + ")")
		return err
	}
	state.ApplicationData = data
	state.Locking = locking
	if err := db.Update(state); err != nil {
		db.PErr.Println("BOLT : Failed (ERR code :" + err.Error() + ")")
		return err
	}
	return nil
}

func (db dbData) Update(data NodeState) error {
	db.pInfo.Println("BOLT : Commence window request update")
	if err := db.database.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Node"))
		var put_Temp []byte
		var err error
		if put_Temp, err = json.Marshal(data); err != nil {
			return fmt.Errorf(data.Identity)
		}
		if err = bucket.Put([]byte(data.Identity), put_Temp); err != nil {
			return fmt.Errorf(data.Identity + ",put error")
		}
		return nil
	}); err != nil {
		return fmt.Errorf("Failed to update window request")
	}

	return nil
}
/*
//Check the window is online 0 : OFFLINE 1 : ONLINE
func (db dbData) IsOnline(key string) (bool, error) {
	db.pInfo.Println("BOLT : State query initiated")
	var state_Temp NodeState

	if err := db.database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Node"))
		data := bucket.Get([]byte(key))
		if err := json.Unmarshal(data, &state_Temp); err != nil {
			db.PErr.Println("BOLT : Unmarshal failed")
			return fmt.Errorf("Unmarshal failed")
		}
		return nil
	}); err != nil {
		return false, err
	}
	return state_Temp.IsOnline, nil
}
*/
func (db dbData) IsRequest(key string) (bool, error) {
	db.pInfo.Println("BOLT : State req query initiated")
	var state_Temp NodeState

	if err := db.database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Node"))
		data := bucket.Get([]byte(key))
		if err := json.Unmarshal(data, &state_Temp); err != nil {
			db.PErr.Println("BOLT ; Unmarshal failed")
			return fmt.Errorf("Unmarshal failed")
		}
		return nil
	}); err != nil {
		return false, err
	}
	return state_Temp.IsRequireConn, nil
}

func (db dbData) GetWindowData(key string) (NodeState, error) {
	db.pInfo.Println("BOLT : Getting state query")
	var node NodeState

	if err := db.database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Node"))
		data := bucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf(NOTAVAILABLE)
		}
		if err := json.Unmarshal(data, &node); err != nil {
			return fmt.Errorf("Unmarshal failed")
		}
		return nil
	}); err != nil {
		return NodeState{}, err
	}

	return node, nil
}
