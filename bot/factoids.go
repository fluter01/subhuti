// Copyright 2016 Alex Fluter

package bot

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrFactoidNotFound = errors.New("Factoid does not exist")
	ErrFactoidExists   = errors.New("Factoid already exist")
)

type Factoid struct {
	Network  string
	Channel  string
	Owner    string
	Nick     string
	Keyword  string
	Desc     string
	Created  time.Time
	RefCount int
	RefUser  string
	Enabled  bool
}

func (f *Factoid) String() string {
	return fmt.Sprintf("%s: %s", f.Keyword, f.Desc)
}

type channelFactoids struct {
	channel  string
	factoids map[string]*Factoid
}

type networkFactoids struct {
	network  string
	channels map[string]*channelFactoids
}

type Factoids struct {
	networks map[string]*networkFactoids
	store    Store
}

func NewFactoids(dbpath string) *Factoids {
	store, err := NewStoreSpace(dbpath, FACTOID)
	if err != nil {
		panic(err)
		return nil
	}

	factoids := new(Factoids)
	factoids.store = store
	factoids.networks = make(map[string]*networkFactoids)

	if err := factoids.load(); err != nil {
		return nil
	}

	return factoids
}

func (factoids *Factoids) Add(fact *Factoid) error {
	if err := factoids.add(fact); err != nil {
		return err
	}
	if err := factoids.saveOne(fact); err != nil {
		return err
	}
	return nil
}

func (factoids *Factoids) add(fact *Factoid) error {
	var (
		network *networkFactoids
		channel *channelFactoids
		factoid *Factoid
		ok      bool
	)

	factoid = new(Factoid)
	*factoid = *fact
	if network, ok = factoids.networks[fact.Network]; !ok {
		network = &networkFactoids{
			network:  fact.Network,
			channels: make(map[string]*channelFactoids),
		}
		channel = &channelFactoids{
			channel:  fact.Channel,
			factoids: make(map[string]*Factoid),
		}
		channel.factoids[fact.Keyword] = factoid
		network.channels[fact.Channel] = channel
		factoids.networks[fact.Network] = network
	} else {
		if channel, ok = network.channels[fact.Channel]; !ok {
			channel = &channelFactoids{
				channel:  fact.Channel,
				factoids: make(map[string]*Factoid),
			}
			channel.factoids[fact.Keyword] = factoid
			network.channels[fact.Channel] = channel
		} else {
			return ErrFactoidExists
		}
	}
	return nil
}

func (factoids *Factoids) Remove(fact *Factoid) error {
	var (
		network *networkFactoids
		channel *channelFactoids
		ok      bool
	)

	if network, ok = factoids.networks[fact.Network]; !ok {
		return ErrFactoidNotFound
	} else {
		if channel, ok = network.channels[fact.Channel]; !ok {
			return ErrFactoidNotFound
		} else {
			if _, ok = channel.factoids[fact.Keyword]; !ok {
				return ErrFactoidNotFound
			} else {
				delete(channel.factoids, fact.Keyword)
			}
		}
	}

	if err := factoids.removeOne(fact); err != nil {
		return err
	}
	return nil
}

func (factoids *Factoids) Dump(w io.Writer) error {
	w.Write([]byte("Factoids\n"))
	for _, ns := range factoids.networks {
		s := fmt.Sprintf("\t%s\n", ns.network)
		w.Write([]byte(s))

		for _, cs := range ns.channels {
			s := fmt.Sprintf("\t\t%s\n", cs.channel)
			w.Write([]byte(s))

			for _, f := range cs.factoids {
				s := fmt.Sprintf("\t\t\t%s\n", f)
				w.Write([]byte(s))
			}
		}
	}

	return nil
}

func (factoids *Factoids) Close() {
	factoids.save()
	factoids.store.Close()
}

func (factoids *Factoids) load() error {
	var (
		key     string
		value   []byte
		pairs   []*Pair
		buf     *bytes.Buffer
		factoid Factoid
		err     error
	)

	pairs, err = factoids.store.List()
	if err != nil {
		return err
	}

	for _, pair := range pairs {
		key = pair.Key
		value = pair.Value
		buf = bytes.NewBuffer(value)
		dec := gob.NewDecoder(buf)
		if err := dec.Decode(&factoid); err != nil {
			return err
		}

		// TODO: validate key
		_ = key

		factoids.add(&factoid)
	}

	return nil
}

func (factoids *Factoids) save() error {
	for _, network := range factoids.networks {
		for _, channel := range network.channels {
			for _, factoid := range channel.factoids {
				if err := factoids.saveOne(factoid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (factoids *Factoids) saveOne(factoid *Factoid) error {
	var (
		key   string
		value []byte
		buf   bytes.Buffer
	)

	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(factoid); err != nil {
		return err
	}

	key = fmt.Sprintf("%s.%s.%s",
		factoid.Network,
		factoid.Channel,
		factoid.Keyword)
	value = buf.Bytes()

	if err := factoids.store.Put(key, value); err != nil {
		return err
	}
	return nil
}

func (factoids *Factoids) removeOne(factoid *Factoid) error {
	var (
		key string
	)

	key = fmt.Sprintf("%s.%s.%s",
		factoid.Network,
		factoid.Channel,
		factoid.Keyword)

	if err := factoids.store.Delete(key); err != nil {
		return err
	}
	return nil
}
