package bot

import (
	"github.com/boltdb/bolt"
)

const (
	Factoid = "FACTOID"
)

type Pair struct {
	Key   string
	Value []byte
}

type Store interface {
	Put(key string, value []byte) error
	Get(key string) (*Pair, error)
	Delete(key string) error
	Exists(key string) (bool, error)
	List() ([]*Pair, error)
	Close()
}

type BoltStore struct {
	db *bolt.DB
}

func NewStore() (Store, error) {
	db, err := bolt.Open("", 0644, nil)
	if err != nil {
		return nil, err
	}
	return &BoltStore{db: db}, nil
}

func (b *BoltStore) Put(key string, value []byte) error {
	return nil
}

func (b *BoltStore) Get(key string) (*Pair, error) {
	return nil, nil
}

func (b *BoltStore) Delete(key string) error {
	return nil
}

func (b *BoltStore) Exists(key string) (bool, error) {
	return false, nil
}

func (b *BoltStore) List() ([]*Pair, error) {
	return nil, nil
}

func (b *BoltStore) Close() {
}
