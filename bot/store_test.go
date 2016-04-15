package bot

import (
	"bytes"
	"testing"
)

const path = "../data/test"

func TestStore(t *testing.T) {
	store, err := NewStore(path)

	if err != nil {
		t.Fatal(err)
	}

	data := map[string][]byte{
		"one":   []byte("111"),
		"two":   []byte("222"),
		"three": []byte("333"),
		"four":  []byte("444"),
	}

	for k, v := range data {
		err = store.Put(k, v)
		if err != nil {
			t.Error(err)
		}

		exists, err := store.Exists(k)
		if err != nil {
			t.Error(err)
		}
		if !exists {
			t.Fail()
		}

		pair, err := store.Get(k)
		if err != nil {
			t.Error(err)
		}
		if bytes.Compare(pair.Value, v) != 0 {
			t.Error(pair.Value, "!=", v)
		}
	}
	exists, err := store.Exists("not exists")
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Fail()
	}

	pairs, err := store.List()
	if err != nil {
		t.Error(err)
	}
	for _, p := range pairs {
		if bytes.Compare(p.Value, data[p.Key]) != 0 {
			t.Fail()
		}
	}

	for k, _ := range data {
		err = store.Delete(k)
		if err != nil {
			t.Error(err)
		}

		exists, err := store.Exists(k)
		if err != nil {
			t.Error(err)
		}
		if exists {
			t.Error()
		}

		pair, err := store.Get(k)
		if err != ErrKeyNotFound {
			t.Error(err)
		}
		if pair != nil {
			t.Fail()
		}
	}
	store.Close()
}
