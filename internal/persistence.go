package internal

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"strconv"
	"time"
)

type Repository interface {
	GetMetadata() (*Metadata, error)
	SaveMetadata(metadata *Metadata) error
	GetLogs() ([]*Log, error)
	SaveLog(index int64, log Log) error
}

type EtcdRepository struct {
	client *clientv3.Client
}

func NewEtcdRepository() (*EtcdRepository, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{os.Getenv("ETCD_ENDPOINT")},
		DialTimeout: time.Second,
	})
	return &EtcdRepository{client: client}, err
}

func (er *EtcdRepository) GetMetadata() (*Metadata, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	resp, err := er.client.Get(ctx, "raft/metadata")
	//cancel()
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{ServerId: uuid.New().String()}

	fmt.Println(resp.Kvs)

	for _, kv := range resp.Kvs {
		switch string(kv.Key) {
		case "raft/metadata/server-id":
			metadata.ServerId = string(kv.Value)
		case "raft/metadata/current-term":
			val, _ := strconv.Atoi(string(kv.Value))
			metadata.CurrentTerm = int64(val)
		case "raft/metadata/voted-for":
			metadata.VotedFor = string(kv.Value)
		case "raft/metadata/commit-index":
			val, _ := strconv.Atoi(string(kv.Value))
			metadata.CommitIndex = int64(val)
		case "raft/metadata/last-applied":
			val, _ := strconv.Atoi(string(kv.Value))
			metadata.LastApplied = int64(val)
		}
	}

	return metadata, nil
}

func (er *EtcdRepository) SaveMetadata(metadata *Metadata) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	resp, err := er.client.Put(ctx, "raft/metadata/server-id", metadata.ServerId)
	cancel()
	fmt.Println(resp)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	resp, err = er.client.Put(ctx, "raft/metadata/current-term", strconv.Itoa(int(metadata.CurrentTerm)))
	cancel()
	fmt.Println(resp)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	_, err = er.client.Put(context.TODO(), "raft/metadata/voted-for", metadata.VotedFor)
	cancel()
	fmt.Println(resp)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	_, err = er.client.Put(context.TODO(), "raft/metadata/commit-index", strconv.Itoa(int(metadata.CommitIndex)))
	cancel()
	fmt.Println(resp)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	_, err = er.client.Put(context.TODO(), "raft/metadata/last-applied", strconv.Itoa(int(metadata.LastApplied)))
	cancel()
	fmt.Println(resp)
	if err != nil {
		return err
	}

	return nil
}

func (er *EtcdRepository) GetLogs() ([]*Log, error) {
	return []*Log{}, nil
}

func (er *EtcdRepository) SaveLog(index int64, log Log) error {
	return nil
}
