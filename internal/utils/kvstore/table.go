// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package kvstore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/iwind/TeaGo/types"
	"sync"
)

const (
	KeyPrefix    = "K$"
	KeyMaxLength = 8 << 10

	FieldPrefix = "F$"

	MaxBatchKeys = 8 << 10 // TODO not implemented
)

type Table[T any] struct {
	name         string
	rawNamespace []byte
	db           *DB
	encoder      ValueEncoder[T]
	fieldNames   []string
	isClosed     bool

	mu *sync.RWMutex
}

func NewTable[T any](tableName string, encoder ValueEncoder[T]) (*Table[T], error) {
	if !IsValidName(tableName) {
		return nil, errors.New("invalid table name '" + tableName + "'")
	}

	return &Table[T]{
		name:    tableName,
		encoder: encoder,
		mu:      &sync.RWMutex{},
	}, nil
}

func (this *Table[T]) Name() string {
	return this.name
}

func (this *Table[T]) Namespace() []byte {
	var dest = make([]byte, len(this.rawNamespace))
	copy(dest, this.rawNamespace)
	return dest
}

func (this *Table[T]) SetNamespace(namespace []byte) {
	this.rawNamespace = namespace
}

func (this *Table[T]) SetDB(db *DB) {
	this.db = db
}

func (this *Table[T]) DB() *DB {
	return this.db
}

func (this *Table[T]) Encoder() ValueEncoder[T] {
	return this.encoder
}

func (this *Table[T]) Set(key string, value T) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	if len(key) > KeyMaxLength {
		return ErrKeyTooLong
	}

	valueBytes, err := this.encoder.Encode(value)
	if err != nil {
		return err
	}

	return this.WriteTx(func(tx *Tx[T]) error {
		return this.set(tx, key, valueBytes, value, false, false)
	})
}

func (this *Table[T]) SetSync(key string, value T) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	if len(key) > KeyMaxLength {
		return ErrKeyTooLong
	}

	valueBytes, err := this.encoder.Encode(value)
	if err != nil {
		return err
	}

	return this.WriteTxSync(func(tx *Tx[T]) error {
		return this.set(tx, key, valueBytes, value, false, true)
	})
}

func (this *Table[T]) Insert(key string, value T) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	if len(key) > KeyMaxLength {
		return ErrKeyTooLong
	}

	valueBytes, err := this.encoder.Encode(value)
	if err != nil {
		return err
	}

	return this.WriteTx(func(tx *Tx[T]) error {
		return this.set(tx, key, valueBytes, value, true, false)
	})
}

// ComposeFieldKey compose field key
// $Namespace$FieldName$FieldValueSeparatorKeyValueFieldLength[2]
func (this *Table[T]) ComposeFieldKey(keyBytes []byte, fieldName string, fieldValueBytes []byte) []byte {
	// TODO use 'make()' and 'copy()' to pre-alloc memory space
	var b = make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(len(fieldValueBytes)))
	var fieldKey = append(this.FieldKey(fieldName), '$') // namespace
	fieldKey = append(fieldKey, fieldValueBytes...)      // field value
	fieldKey = append(fieldKey, 0, 0)                    // separator
	fieldKey = append(fieldKey, keyBytes...)             // key value
	fieldKey = append(fieldKey, b...)                    // field value length
	return fieldKey
}

func (this *Table[T]) Exist(key string) (found bool, err error) {
	if this.isClosed {
		return false, NewTableClosedErr(this.name)
	}

	_, closer, err := this.db.store.rawDB.Get(this.FullKey(key))
	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	defer func() {
		_ = closer.Close()
	}()

	return true, nil
}

func (this *Table[T]) Get(key string) (value T, err error) {
	if this.isClosed {
		err = NewTableClosedErr(this.name)
		return
	}

	err = this.ReadTx(func(tx *Tx[T]) error {
		resultValue, getErr := this.get(tx, key)
		if getErr == nil {
			value = resultValue
		}
		return getErr
	})

	return
}

func (this *Table[T]) Delete(key ...string) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	if len(key) == 0 {
		return nil
	}

	return this.WriteTx(func(tx *Tx[T]) error {
		return this.deleteKeys(tx, key...)
	})
}

func (this *Table[T]) ReadTx(fn func(tx *Tx[T]) error) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	tx, err := NewTx[T](this, true)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Close()
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (this *Table[T]) WriteTx(fn func(tx *Tx[T]) error) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	tx, err := NewTx[T](this, false)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Close()
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (this *Table[T]) WriteTxSync(fn func(tx *Tx[T]) error) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	tx, err := NewTx[T](this, false)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Close()
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.CommitSync()
}

func (this *Table[T]) Truncate() error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	return this.db.store.rawDB.DeleteRange(this.Namespace(), append(this.Namespace(), 0xFF), DefaultWriteOptions)
}

func (this *Table[T]) DeleteRange(start string, end string) error {
	if this.isClosed {
		return NewTableClosedErr(this.name)
	}

	return this.db.store.rawDB.DeleteRange(this.FullKeyBytes([]byte(start)), this.FullKeyBytes([]byte(end)), DefaultWriteOptions)
}

func (this *Table[T]) Query() *Query[T] {
	var query = NewQuery[T]()
	query.SetTable(this)
	return query
}

func (this *Table[T]) Count() (int64, error) {
	var count int64

	var begin = this.FullKeyBytes(nil)
	it, err := this.db.store.rawDB.NewIter(&pebble.IterOptions{
		LowerBound: begin,
		UpperBound: append(begin, 0xFF),
	})
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = it.Close()
	}()

	for it.First(); it.Valid(); it.Next() {
		count++
	}

	return count, err
}

func (this *Table[T]) FullKey(realKey string) []byte {
	return append(this.Namespace(), KeyPrefix+realKey...)
}

func (this *Table[T]) FullKeyBytes(realKeyBytes []byte) []byte {
	var k = append(this.Namespace(), KeyPrefix...)
	k = append(k, realKeyBytes...)
	return k
}

func (this *Table[T]) FieldKey(fieldName string) []byte {
	var data = append(this.Namespace(), FieldPrefix...)
	data = append(data, fieldName...)
	return data
}

func (this *Table[T]) DecodeFieldKey(fieldName string, fieldKey []byte) (fieldValue []byte, key []byte, err error) {
	var l = len(fieldKey)
	var baseLen = len(this.FieldKey(fieldName)) + 1 /** $ **/ + 2 /** separator length **/ + 2 /** field length **/
	if l < baseLen {
		err = errors.New("invalid field key")
		return
	}

	var fieldValueLen = binary.BigEndian.Uint16(fieldKey[l-2:])
	var data = fieldKey[baseLen-4 : l-2]

	fieldValue = data[:fieldValueLen]
	key = data[fieldValueLen+2: /** separator length **/]

	return
}

func (this *Table[T]) Close() error {
	this.isClosed = true
	return nil
}

func (this *Table[T]) deleteKeys(tx *Tx[T], key ...string) error {
	var batch = tx.batch

	for _, singleKey := range key {
		var keyErr = func(singleKey string) error {
			var keyBytes = this.FullKey(singleKey)

			// delete field values
			if len(this.fieldNames) > 0 {
				valueBytes, closer, getErr := batch.Get(keyBytes)
				if getErr != nil {
					if IsNotFound(getErr) {
						return nil
					}
					return getErr
				}
				defer func() {
					_ = closer.Close()
				}()

				value, decodeErr := this.encoder.Decode(valueBytes)
				if decodeErr != nil {
					return fmt.Errorf("decode value failed: %w", decodeErr)
				}

				for _, fieldName := range this.fieldNames {
					fieldValueBytes, fieldErr := this.encoder.EncodeField(value, fieldName)
					if fieldErr != nil {
						return fieldErr
					}

					deleteKeyErr := batch.Delete(this.ComposeFieldKey([]byte(singleKey), fieldName, fieldValueBytes), DefaultWriteOptions)
					if deleteKeyErr != nil {
						return deleteKeyErr
					}
				}
			}

			err := batch.Delete(keyBytes, DefaultWriteOptions)
			if err != nil {
				return err
			}

			return nil
		}(singleKey)
		if keyErr != nil {
			return keyErr
		}
	}

	return nil
}

func (this *Table[T]) set(tx *Tx[T], key string, valueBytes []byte, value T, insertOnly bool, syncMode bool) error {
	var keyBytes = this.FullKey(key)
	var writeOptions = DefaultWriteOptions
	if syncMode {
		writeOptions = DefaultWriteSyncOptions
	}

	var batch = tx.batch

	// read old value
	var oldValue T
	var oldFound bool
	var countFields = len(this.fieldNames)

	if !insertOnly {
		if countFields > 0 {
			oldValueBytes, closer, getErr := batch.Get(keyBytes)
			if getErr != nil {
				if !IsNotFound(getErr) {
					return getErr
				}
			} else {
				defer func() {
					_ = closer.Close()
				}()

				var decodeErr error
				oldValue, decodeErr = this.encoder.Decode(oldValueBytes)
				if decodeErr != nil {
					return fmt.Errorf("decode value failed: %w", decodeErr)
				}
				oldFound = true
			}
		}
	}

	setErr := batch.Set(keyBytes, valueBytes, writeOptions)
	if setErr != nil {
		return setErr
	}

	// process fields
	if countFields > 0 {
		// add new field keys
		for _, fieldName := range this.fieldNames {
			// 把EncodeField放在TX里，是为了节约内存
			fieldValueBytes, fieldErr := this.encoder.EncodeField(value, fieldName)
			if fieldErr != nil {
				return fieldErr
			}

			if len(fieldValueBytes) > 8<<10 {
				return errors.New("field value too long: " + types.String(len(fieldValueBytes)))
			}

			var newFieldKeyBytes = this.ComposeFieldKey([]byte(key), fieldName, fieldValueBytes)

			// delete old field key
			if oldFound {
				oldFieldValueBytes, oldFieldErr := this.encoder.EncodeField(oldValue, fieldName)
				if oldFieldErr != nil {
					return oldFieldErr
				}
				var oldFieldKeyBytes = this.ComposeFieldKey([]byte(key), fieldName, oldFieldValueBytes)
				if bytes.Equal(oldFieldKeyBytes, newFieldKeyBytes) {
					// skip the field
					continue
				}
				deleteFieldErr := batch.Delete(oldFieldKeyBytes, writeOptions)
				if deleteFieldErr != nil {
					return deleteFieldErr
				}
			}

			// set new field key
			setFieldErr := batch.Set(newFieldKeyBytes, nil, writeOptions)
			if setFieldErr != nil {
				return setFieldErr
			}
		}
	}

	return nil
}

func (this *Table[T]) get(tx *Tx[T], key string) (value T, err error) {
	return this.getWithKeyBytes(tx, this.FullKey(key))
}

func (this *Table[T]) getWithKeyBytes(tx *Tx[T], keyBytes []byte) (value T, err error) {
	valueBytes, closer, err := tx.batch.Get(keyBytes)
	if err != nil {
		return value, err
	}
	defer func() {
		_ = closer.Close()
	}()

	resultValue, decodeErr := this.encoder.Decode(valueBytes)
	if decodeErr != nil {
		return value, fmt.Errorf("decode value failed: %w", decodeErr)
	}
	value = resultValue
	return
}
