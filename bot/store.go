package bot

import (
	"errors"

	"github.com/boltdb/bolt"
)

type StoreSpace int

const (
	ROOT StoreSpace = iota
	FACTOID
)

const (
	root    = "ROOT"
	factoid = "FACTOID"
)

var (
	SpaceNames = map[StoreSpace]string{
		ROOT:    root,
		FACTOID: factoid,
	}
)

var (
	ErrDirNotFound = errors.New("Directory not found")
	ErrKeyNotFound = errors.New("Key not found")
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

func NewStore(path string) (Store, error) {
	db, err := bolt.Open(path, 0644, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range SpaceNames {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &BoltStore{db: db}, nil
}

func (b *BoltStore) Put(key string, value []byte) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(root))
		if bucket == nil {
			return ErrDirNotFound
		}
		err := bucket.Put([]byte(key), value)

		return err
	})
	return err
}

func (b *BoltStore) Get(key string) (*Pair, error) {
	var val []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(root))
		if bucket == nil {
			return ErrDirNotFound
		}
		v := bucket.Get([]byte(key))
		if v != nil {
			val = dup(v)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, ErrKeyNotFound
	}
	return &Pair{key, val}, nil
}

func (b *BoltStore) Delete(key string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(root))
		if bucket == nil {
			return ErrDirNotFound
		}
		err := bucket.Delete([]byte(key))

		return err
	})
	return err
}

func (b *BoltStore) Exists(key string) (bool, error) {
	var exists bool
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(root))
		if bucket == nil {
			return ErrDirNotFound
		}
		v := bucket.Get([]byte(key))
		if v != nil {
			exists = true
		}

		return nil
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (b *BoltStore) List() ([]*Pair, error) {
	pairs := make([]*Pair, 0)
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(root))
		if bucket == nil {
			return ErrDirNotFound
		}
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			val := dup(v)
			pairs = append(pairs, &Pair{string(k), val})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (b *BoltStore) Close() {
	b.db.Close()
}

func dup(src []byte) []byte {
	new := make([]byte, len(src))
	copy(new, src)
	return new
}
