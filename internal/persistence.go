package internal

import (
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"strconv"
	"time"
)

type Repository interface {
	GetMetadata() (*Metadata, error)
	SaveMetadata(metadata *Metadata) error
	GetLogs() ([]*Log, error)
	SaveLog(index int64, log Log) error
}

type BoltRepository struct {
	db *bolt.DB
}

func NewBoltRepository(path string) (*BoltRepository, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	//defer db.Close()
	return &BoltRepository{db: db}, nil
}

func (bp *BoltRepository) Cleanup() {
	bp.db.Close()
}

func GetStringValueFromBucket(key string, bucket *bolt.Bucket) (string, error) {
	value := bucket.Get([]byte(key))

	if value == nil {
		err := bucket.Put([]byte(key), []byte(""))
		return "", err
	}

	return string(value), nil
}

func GetInt64ValueFromBucket(key string, bucket *bolt.Bucket) (int64, error) {
	value := bucket.Get([]byte(key))

	if value == nil {
		err := bucket.Put([]byte(key), []byte("0"))
		return 0, err
	}

	number, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, err
	}

	return int64(number), nil
}

func (bp *BoltRepository) GetMetadata() (*Metadata, error) {
	var metadata Metadata

	err := bp.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("RaftMetadata"))
		if err != nil {
			return err
		}

		serverId := bucket.Get([]byte("serverId"))

		if serverId != nil {
			metadata.ServerId = string(serverId)
		} else {
			newUuid, err := uuid.NewUUID()
			if err != nil {
				return err
			}

			err = bucket.Put([]byte("serverId"), []byte(newUuid.String()))
			if err != nil {
				return err
			}
			metadata.ServerId = newUuid.String()
		}

		metadata.CurrentTerm, err = GetInt64ValueFromBucket("currentTerm", bucket)
		if err != nil {
			return err
		}

		metadata.VotedFor, err = GetStringValueFromBucket("votedFor", bucket)
		if err != nil {
			return err
		}

		metadata.CommitIndex, err = GetInt64ValueFromBucket("commitIndex", bucket)
		if err != nil {
			return err
		}

		metadata.LastApplied, err = GetInt64ValueFromBucket("lastApplied", bucket)
		if err != nil {
			return err
		}

		return nil
	})

	return &metadata, err
}

func (bp *BoltRepository) SaveMetadata(metadata *Metadata) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("RaftMetadata"))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("serverId"), []byte(metadata.ServerId))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("currentTerm"), []byte(strconv.Itoa(int(metadata.CurrentTerm))))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("votedFor"), []byte(metadata.VotedFor))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("commitIndex"), []byte(strconv.Itoa(int(metadata.CommitIndex))))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("lastApplied"), []byte(strconv.Itoa(int(metadata.LastApplied))))
		if err != nil {
			return err
		}

		return nil
	})
}

func (bp *BoltRepository) GetLogs() ([]*Log, error) {
	return nil, nil
}

func (bp *BoltRepository) SaveLog(index int64, log Log) error {
	return nil
}
