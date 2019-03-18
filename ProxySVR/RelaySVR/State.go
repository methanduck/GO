package main

import (
	"fmt"
	bolt2 "github.com/boltdb/bolt"
	"log"
	"strconv"
)

type dbData struct {
	database *bolt2.DB
	pInfo    *log.Logger
	PErr     *log.Logger
}

func (db dbData) Startbolt(pinfo *log.Logger, perr *log.Logger) {
	db.pInfo = pinfo
	db.PErr = perr
	bolt, err := bolt2.Open("SmartWindow.db", 0600, nil)
	if err != nil {
		log.Panic("Failed to open bolt database")
	}
	db.database = bolt
	defer func() {
		if err := bolt.Close(); err != nil {
			perr.Panic("bolt database abnormally terminated")
		}
	}()

	pinfo.Println("BOLT : create new bucket")
	bolt.Update(func(tx *bolt2.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("Node")); err != nil {
			return fmt.Errorf("BOLT : bucket creation failed")
		}
		pinfo.Println("BOLT : bucket creation secceded")
		return nil
	})
}

func (db dbData) Update(key string, val bool) {
	pInfo.Println("BOLT : Update prcessing")
	var tmp string
	db.database.Update(func(tx *bolt2.Tx) error {
		bucket := tx.Bucket([]byte("Node"))
		if val {
			tmp = "1"
		} else {
			tmp = "0"
		}
		if err := bucket.Put([]byte(key), []byte(tmp)); err != nil {
			pErr.Println("BOLT : failed to update database (Err code, key :" + key + "val :" + strconv.FormatBool(val))
		}
		return nil
	})
}
