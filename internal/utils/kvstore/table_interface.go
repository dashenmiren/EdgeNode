package kvstore

type TableInterface interface {
	Name() string
	SetNamespace(namespace []byte)
	SetDB(db *DB)
	Close() error
}
