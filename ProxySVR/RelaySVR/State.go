package RelaySVR

import (
	"encoding/json"
	"fmt"
	"github.com/GO/InteractiveSocket"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

const (
	ONLINE   = true
	OFFLINE  = false
	NOTFOUND = "N/A"
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
	if err := boltdb.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("Node")); err != nil {
			return fmt.Errorf("bucket creation failed")
		}
		pinfo.Println("BOLT : bucket creation secceded")
		return nil
	}); err != nil {
		db.PErr.Println("BOLT : Failed Startbolt (ERR code :" + err.Error() + ")")
	}
	return boltdb
}

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
			return fmt.Errorf(NOTFOUND)
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
