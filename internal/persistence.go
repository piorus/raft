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
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("raft"))
		return err
	})

	//defer db.Close()
	return &BoltRepository{db: db}, err
}

func (bp *BoltRepository) Cleanup() {
	bp.db.Close()
}

func (bp *BoltRepository) GetMetadata() (*Metadata, error) {
	newUuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	metadata := &Metadata{ServerId: newUuid.String()}

	err = bp.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("raft"))
		cursor := bucket.Cursor()
		k, v := cursor.Seek([]byte("metadata/"))
		for {
			if k == nil {
				break
			}
			switch string(k) {
			case "metadata/server-id":
				metadata.ServerId = string(v)
			case "metadata/current-term":
				val, err := strconv.Atoi(string(v))
				if err != nil {
					return err
				}
				metadata.CurrentTerm = int64(val)
			case "metadata/voted-for":
				metadata.VotedFor = string(v)
			case "metadata/commit-index":
				val, err := strconv.Atoi(string(v))
				if err != nil {
					return err
				}
				metadata.CommitIndex = int64(val)
			case "metadata/last-applied":
				val, err := strconv.Atoi(string(v))
				if err != nil {
					return err
				}
				metadata.LastApplied = int64(val)
			}
			k, v = cursor.Next()
		}

		return nil
	})

	return metadata, err
}

func (bp *BoltRepository) SaveMetadata(metadata *Metadata) error {
	return bp.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("raft"))

		err := bucket.Put([]byte("metadata/server-id"), []byte(metadata.ServerId))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("metadata/current-term"), []byte(strconv.Itoa(int(metadata.CurrentTerm))))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("metadata/voted-for"), []byte(metadata.VotedFor))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("metadata/commit-index"), []byte(strconv.Itoa(int(metadata.CommitIndex))))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("metadata/last-applied"), []byte(strconv.Itoa(int(metadata.LastApplied))))
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
