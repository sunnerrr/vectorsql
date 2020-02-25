// Copyright 2019 The OctoSQL Authors.
// Copyright 2020 The VectorSQL Authors.
//
// Code is licensed under Apache License, Version 2.0.

package datavalues

import (
	"github.com/segmentio/fasthash/fnv1a"
)

type HashMap struct {
	count     int
	container map[uint64][]entry
}

func NewHashMap() *HashMap {
	return &HashMap{
		container: make(map[uint64][]entry),
	}
}

type entry struct {
	key   IDataValue
	value interface{}
}

func AreEqual(left, right IDataValue) bool {
	if left.GetType() != right.GetType() {
		return false
	}
	cmp, err := left.Compare(right)
	if err != nil {
		return false
	}
	return cmp == Equal
}

func fastHash(key IDataValue) uint64 {
	h2 := fnv1a.Init64
	h2 = fnv1a.AddString64(h2, (key.Show()))
	return h2
}

func (hm *HashMap) Set(key IDataValue, value interface{}) error {
	hash := fastHash(key)
	list := hm.container[hash]
	for i := range list {
		if AreEqual(list[i].key, key) {
			list[i].value = value
			return nil
		}
	}
	hm.container[hash] = append(list, entry{
		key:   key,
		value: value,
	})
	hm.count++
	return nil
}

func (hm *HashMap) SetByHash(key IDataValue, hash uint64, value interface{}) error {
	list := hm.container[hash]
	hm.container[hash] = append(list, entry{
		key:   key,
		value: value,
	})
	hm.count++
	return nil
}

func (hm *HashMap) Get(key IDataValue) (interface{}, uint64, bool, error) {
	hash := fastHash(key)
	list := hm.container[hash]
	for i := range list {
		if AreEqual(list[i].key, key) {
			return list[i].value, hash, true, nil
		}
	}
	return nil, hash, false, nil
}

func (hm *HashMap) Count() int {
	return hm.count
}

func (hm *HashMap) GetIterator() *HashMapIterator {
	hashes := make([]uint64, 0, len(hm.container))
	for k := range hm.container {
		hashes = append(hashes, k)
	}

	return &HashMapIterator{
		hm:             hm,
		hashes:         hashes,
		hashesPosition: 0,
		listPosition:   0,
	}
}

type HashMapIterator struct {
	hm             *HashMap
	hashes         []uint64
	hashesPosition int
	listPosition   int
}

func (iter *HashMapIterator) Next() (IDataValue, interface{}, bool) {
	if iter.hashesPosition == len(iter.hashes) {
		return nil, nil, false
	}

	// Save current item location
	outHashPos := iter.hashesPosition
	outListPos := iter.listPosition

	// Advance iterator to next item
	if iter.listPosition+1 == len(iter.hm.container[iter.hashes[iter.hashesPosition]]) {
		iter.hashesPosition++
		iter.listPosition = 0
	} else {
		iter.listPosition++
	}

	outEntry := iter.hm.container[iter.hashes[outHashPos]][outListPos]
	return outEntry.key, outEntry.value, true
}
