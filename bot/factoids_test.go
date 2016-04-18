package bot

import (
	"testing"
	"time"
)

func TestFactoids(t *testing.T) {
	fs := NewFactoids(testdbpath)

	if fs == nil {
		t.Log("cannot get factoids")
		t.Fatal()
	}
	fact := &Factoid{
		Network:  "testnet",
		Channel:  "##candice",
		Owner:    G,
		Nick:     G,
		Keyword:  "hi2",
		Desc:     "hello",
		Created:  time.Now(),
		RefCount: 0,
		RefUser:  "none",
		Enabled:  true,
	}

	err := fs.Add(fact)
	if err != nil {
		t.Error(err)
	}

	fact.Keyword = "moo2"
	fact.Desc = "cow"
	err = fs.Add(fact)
	if err != nil {
		t.Error(err)
	}

	fs.Close()
}
