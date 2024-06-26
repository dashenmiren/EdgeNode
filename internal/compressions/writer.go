package compressions

import "io"

type Writer interface {
	Write(p []byte) (int, error)
	Flush() error
	Reset(writer io.Writer)
	RawClose() error
	Close() error
	Level() int

	SetPool(pool *WriterPool)
	ResetFinish()
}
