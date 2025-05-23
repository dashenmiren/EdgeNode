// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package kvstore

import (
	"errors"
	"github.com/cockroachdb/pebble"
	"sync"
)

type DB struct {
	store *Store

	name      string
	namespace string
	tableMap  map[string]TableInterface

	mu sync.RWMutex
}

func NewDB(store *Store, dbName string) (*DB, error) {
	if !IsValidName(dbName) {
		return nil, errors.New("invalid database name '" + dbName + "'")
	}

	return &DB{
		store:     store,
		name:      dbName,
		namespace: "$" + dbName + "$",
		tableMap:  map[string]TableInterface{},
	}, nil
}

func (this *DB) AddTable(table TableInterface) {
	table.SetNamespace([]byte(this.Namespace() + table.Name() + "$"))
	table.SetDB(this)

	this.mu.Lock()
	defer this.mu.Unlock()

	this.tableMap[table.Name()] = table
}

func (this *DB) Name() string {
	return this.name
}

func (this *DB) Namespace() string {
	return this.namespace
}

func (this *DB) Store() *Store {
	return this.store
}

func (this *DB) Inspect(fn func(key []byte, value []byte)) error {
	it, err := this.store.rawDB.NewIter(&pebble.IterOptions{
		LowerBound: []byte(this.namespace),
		UpperBound: append([]byte(this.namespace), 0xFF, 0xFF),
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = it.Close()
	}()

	for it.First(); it.Valid(); it.Next() {
		value, valueErr := it.ValueAndErr()
		if valueErr != nil {
			return valueErr
		}
		fn(it.Key(), value)
	}

	return nil
}

// Truncate the database
func (this *DB) Truncate() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	var start = []byte(this.Namespace())
	return this.store.rawDB.DeleteRange(start, append(start, 0xFF), DefaultWriteOptions)
}

func (this *DB) Close() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	var lastErr error
	for _, table := range this.tableMap {
		err := table.Close()
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}
