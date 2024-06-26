package writers

import "io"

type BytesCounterWriter struct {
	writer io.Writer
	count  int64
}

func NewBytesCounterWriter(rawWriter io.Writer) *BytesCounterWriter {
	return &BytesCounterWriter{writer: rawWriter}
}

func (this *BytesCounterWriter) RawWriter() io.Writer {
	return this.writer
}

func (this *BytesCounterWriter) Write(p []byte) (n int, err error) {
	n, err = this.writer.Write(p)
	this.count += int64(n)
	return
}

func (this *BytesCounterWriter) Close() error {
	return nil
}

func (this *BytesCounterWriter) TotalBytes() int64 {
	return this.count
}
