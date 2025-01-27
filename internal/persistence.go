package internal

import (
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"time"
)

type DatabaseConn interface {
	GetMetadata() (*Metadata, error)
	SaveMetadata(metadata *Metadata) error
	GetLogs() ([]*Log, error)
	SaveLog(index int64, log Log) error
}

type BoltDatabaseConn struct {
	db         *bolt.DB
	bucketName string
}

func NewBoltDatabaseConn(path string) (*BoltDatabaseConn, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	//defer db.Close()
	return &BoltDatabaseConn{db: db, bucketName: "raft"}, nil
}

func (bp *BoltDatabaseConn) GetMetadata() (*Metadata, error) {
	return nil, nil
}

func (bp *BoltDatabaseConn) SaveMetadata(metadata *Metadata) error {
	return nil
}

func (bp *BoltDatabaseConn) GetLogs() ([]*Log, error) {
	return nil, nil
}

func (bp *BoltDatabaseConn) SaveLog(index int64, log Log) error {
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
