package compressions

import "io"

type Reader interface {
	Read(p []byte) (n int, err error)
	Reset(reader io.Reader) error
	RawClose() error
	Close() error

	SetPool(pool *ReaderPool)
	ResetFinish()
}
