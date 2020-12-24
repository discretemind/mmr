package db

import (
	"errors"
	"fmt"
)

type memoryKey [32]byte

type memoryDb struct {
	data map[memoryKey][]byte
}

//Memory memory KV database
func Memory() IDatabase {
	res := &memoryDb{
		data: map[memoryKey][]byte{},
	}
	return res
}

func getKey(id []byte) (res memoryKey) {
	copy(res[:], id[:])
	return
}
func (d *memoryDb) Has(id []byte) bool {
	_, ok := d.data[getKey(id)]
	return ok
}
func (d *memoryDb) Get(id []byte) (data []byte, err error) {
	data, ok := d.data[getKey(id)]
	if !ok {
		return nil, errors.New("Not found")
	}
	return data, nil
}

func (d *memoryDb) Set(id []byte, data []byte) error {
	d.data[getKey(id)] = data
	return nil
}

func (d *memoryDb) Dump() string {
	res := ""
	for k, v := range d.data {
		res += fmt.Sprintf("%x : %x\n", k, v)
	}
	return res
}
