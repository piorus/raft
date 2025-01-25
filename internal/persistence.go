package internal

import (
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

type DatabaseConn interface {
	CommitLog(log Log) error
	GetServerId() (string, error)
}

type BoltDatabaseConn struct {
	db         *bolt.DB
	bucketName string
}

func NewBoltDatabaseConn(path string) (*BoltDatabaseConn, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	//defer dbConn.Close()
	return &BoltDatabaseConn{db: db, bucketName: "raft"}, nil
}

func (bp *BoltDatabaseConn) CommitLog(log Log) error {
	return nil
}

func (bp *BoltDatabaseConn) GetServerId() (string, error) {
	var serverId string

	err := bp.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bp.bucketName))
		if err != nil {
			return err
		}

		value := bucket.Get([]byte("serverId"))

		if value != nil {
			serverId = string(value)
			return nil
		}

		newUuid, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("serverId"), []byte(newUuid.String()))
		if err != nil {
			return err
		}
		serverId = newUuid.String()

		return nil
	})

	return serverId, err
}
