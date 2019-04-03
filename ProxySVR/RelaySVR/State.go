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
	ONLINE        = true
	OFFLINE       = false
	NOTAVAILABLE  = "N/A"
	ERRPROCESSING = "PSERR"
	BUCKET_NODE   = "NODE"
	//update options
	UPDATE_ONLIE   = 0
	UPDATE_REQCONN = 1
	UPDATE_ALL     = 2
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
	Locking         int //TODO : embedded locking 전환 필요
}

//기본적으로 온라인인 상태로 리턴
func NodeStateMaker(data InteractiveSocket.Node, state NodeState) NodeState {
	return NodeState{ApplicationData: data, Identity: data.Identity, IsOnline: ONLINE, IsRequireConn: state.IsRequireConn, Locking: 0}
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
//bool : 온라인이든 아니든 리턴이 이뤄질 경우 등록은 되어있는것으로 판단 가능
func (db dbData) IsExistAndIsOnline(identity string) (bool, error) {
	var isOnline bool
	tempNodeState := NodeState{}
	if err := db.database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NODE))
		val := bucket.Get([]byte(identity))
		if len(val) == 0 {
			return fmt.Errorf(NOTAVAILABLE)
		}
		if err := json.Unmarshal(val, tempNodeState); err != nil {
			return fmt.Errorf("BOLT : failed to unmarshal")
		}
		isOnline = tempNodeState.IsOnline
		return nil
	}); err != nil {
		if err.Error() == NOTAVAILABLE {
			return false, errors.New(NOTAVAILABLE)
		}
		return false, errors.New(ERRPROCESSING)
	} else {
		return isOnline, nil
	}
}
func (db dbData) IsRequireConn(identity string) (res bool, resErr error) {
	if err := db.database.View(func(tx *bolt.Tx) error {
		var tmp_NodeState NodeState
		bucket := tx.Bucket([]byte(BUCKET_NODE))
		val := bucket.Get([]byte(identity))
		if len(val) == 0 {
			res = false
			resErr = errors.New("BOLT : Data not found")
		}
		_ = json.Unmarshal(val, tmp_NodeState) //에러 처리 안함
		res = tmp_NodeState.IsOnline
		resErr = nil
		return nil
	}); err != nil {
		return
	}
	return
}
func (db dbData) GetNodeData(identity string) NodeState {

}

//Update the Node "State" as Online
func (db dbData) UpdataOnline(data InteractiveSocket.Node) error {
	var val []byte
	_, err := db.IsExistAndIsOnline(data.Identity)
	if err != nil { //err리턴 시 데이터가 존재하지 않거나 marshal 에러 발생
		temp_NodeState := NodeState{}
		if err := db.database.Update(func(tx *bolt.Tx) error {
			temp_NodeState.ApplicationData = data
			temp_NodeState.IsOnline = true
			temp_NodeState.Identity = data.Identity

			bucket := tx.Bucket([]byte(BUCKET_NODE))
			val, _ = json.Marshal(temp_NodeState)
			if err := bucket.Put([]byte(data.Identity), val); err != nil {
				return errors.New("BOLT : Failed to put data")
			}
			return nil
		}); err != nil {
			return err
		}
	} else { //리턴이 존재하므로 데이터는 존재함
		var temp_NodeState NodeState
		if err := db.database.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(BUCKET_NODE))
			val = bucket.Get([]byte(data.Identity))
			if err := json.Unmarshal(val, &temp_NodeState); err != nil {
				return fmt.Errorf("BOLT : Failed to unmarshal")
			}
			//Modify state
			temp_NodeState.IsOnline = true
			//marshalling
			val, err := json.Marshal(temp_NodeState)
			if err != nil {
				return fmt.Errorf("BOLT : Failed to marshal")
			}
			//Update Node
			if err := bucket.Put([]byte(data.Identity), val); err != nil {
				fmt.Println("BOLT : Failed to update NodeState")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

//Update the Node data
func (db dbData) UpdateNodeDataState(data InteractiveSocket.Node, isOnline bool, isRequireConn bool, lock int, opts int) error {
	tmpNodeState := NodeState{}
	var val []byte
	if _, err := db.IsExistAndIsOnline(data.Identity); err != nil { //데이터가 존재하지 않을경우
		tmpNodeState.Identity = data.Identity
		tmpNodeState.IsOnline = isOnline
		tmpNodeState.IsRequireConn = isRequireConn
		tmpNodeState.Locking = lock
		if val, err := json.Marshal(tmpNodeState); err != nil {
			return errors.New("BOLT : Failed to marshal")
		} else {
			if err := db.database.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(BUCKET_NODE))
				if err := bucket.Put([]byte(data.Identity), val); err != nil {
					return errors.New("BOLT : Failed to put data")
				}
				return nil
			}); err != nil {
				return err
			}
		}
	} else { //데이터가 존재할 경우
		var tmp_NodeState NodeState
		if err := db.database.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(BUCKET_NODE))
			val = bucket.Get([]byte(data.Identity))
			_ = json.Unmarshal(val, tmp_NodeState) //에러 처리 하지 않음
			if tmp_NodeState.Locking == 1 {
				return errors.New("BOLT : Resource N/A")
			}
			switch opts {
			case UPDATE_ALL:
				tmp_NodeState.IsOnline = isOnline
				tmp_NodeState.IsRequireConn = isRequireConn
				tmp_NodeState.Locking = lock
			case UPDATE_ONLIE:
				tmp_NodeState.IsOnline = isOnline
			case UPDATE_REQCONN:
				tmp_NodeState.IsRequireConn = isRequireConn
			}
			val, _ = json.Marshal(tmp_NodeState) //에러 처리 하지 않음
			if err := bucket.Put([]byte(data.Identity), val); err != nil {
				return errors.New("BOLT : faile to update data")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

//Reset default state //default isOnline = false , isReqConn = false
func (db dbData) ResetState(identity string, isonline bool, isreqConn bool, lock int) error {
	tmpNodeState := db.GetNodeData(identity)
	if isonline {
		tmpNodeState.IsOnline = true
	} else {
		tmpNodeState.IsOnline = false
	}
	if isreqConn {
		tmpNodeState.IsRequireConn = true
	} else {
		tmpNodeState.IsRequireConn = false
	}
	if lock == 1 {
		tmpNodeState.Locking = 1
	} else {
		tmpNodeState.Locking = 0
	}
	tmpNodeState.ApplicationData = InteractiveSocket.Node{}
	//put resetted struct
	if err := db.database.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NODE))
		if val, err := json.Marshal(tmpNodeState); err != nil {
			return errors.New("BOLT : Marshal failed")
		} else {
			if err := bucket.Put([]byte(identity), val); err != nil {
				return errors.New("BOLT : Put failed")
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
