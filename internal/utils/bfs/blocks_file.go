// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package bfs

import (
	"errors"
	"fmt"
	"github.com/dashenmiren/EdgeNode/internal/utils/zero"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const BFileExt = ".b"

type BlockType string

const (
	BlockTypeHeader BlockType = "header"
	BlockTypeBody   BlockType = "body"
)

type BlocksFile struct {
	opt   *BlockFileOptions
	fp    *os.File
	mFile *MetaFile

	isClosing bool
	isClosed  bool

	mu *sync.RWMutex

	writtenBytes   int64
	writingFileMap map[string]zero.Zero // hash => Zero
	syncAt         time.Time

	readerPool chan *FileReader
	countRefs  int32
}

func NewBlocksFileWithRawFile(fp *os.File, options *BlockFileOptions) (*BlocksFile, error) {
	options.EnsureDefaults()

	var bFilename = fp.Name()
	if !strings.HasSuffix(bFilename, BFileExt) {
		return nil, errors.New("filename '" + bFilename + "' must has a '" + BFileExt + "' extension")
	}

	var mu = &sync.RWMutex{}

	var mFilename = strings.TrimSuffix(bFilename, BFileExt) + MFileExt
	mFile, err := OpenMetaFile(mFilename, mu)
	if err != nil {
		_ = fp.Close()
		return nil, fmt.Errorf("load '%s' failed: %w", mFilename, err)
	}

	AckReadThread()
	_, err = fp.Seek(0, io.SeekEnd)
	ReleaseReadThread()
	if err != nil {
		_ = fp.Close()
		_ = mFile.Close()
		return nil, err
	}

	return &BlocksFile{
		fp:             fp,
		mFile:          mFile,
		mu:             mu,
		opt:            options,
		syncAt:         time.Now(),
		readerPool:     make(chan *FileReader, 32),
		writingFileMap: map[string]zero.Zero{},
	}, nil
}

func OpenBlocksFile(filename string, options *BlockFileOptions) (*BlocksFile, error) {
	// TODO 考虑是否使用flock锁定，防止多进程写冲突
	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			var dir = filepath.Dir(filename)
			_ = os.MkdirAll(dir, 0777)

			// try again
			fp, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
		}

		if err != nil {
			return nil, fmt.Errorf("open blocks file failed: %w", err)
		}
	}

	return NewBlocksFileWithRawFile(fp, options)
}

func (this *BlocksFile) Filename() string {
	return this.fp.Name()
}

func (this *BlocksFile) Write(hash string, blockType BlockType, b []byte, originOffset int64) (n int, err error) {
	if len(b) == 0 {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	posBefore, err := this.currentPos()
	if err != nil {
		return 0, err
	}

	err = this.checkStatus()
	if err != nil {
		return
	}

	AckWriteThread()
	n, err = this.fp.Write(b)
	ReleaseWriteThread()

	if err == nil {
		if n > 0 {
			this.writtenBytes += int64(n)
		}

		if blockType == BlockTypeHeader {
			err = this.mFile.WriteHeaderBlockUnsafe(hash, posBefore, posBefore+int64(n))
		} else if blockType == BlockTypeBody {
			err = this.mFile.WriteBodyBlockUnsafe(hash, posBefore, posBefore+int64(n), originOffset, originOffset+int64(n))
		} else {
			err = errors.New("invalid block type '" + string(blockType) + "'")
		}
	}

	return
}

func (this *BlocksFile) OpenFileWriter(fileHash string, bodySize int64, isPartial bool) (writer *FileWriter, err error) {
	err = CheckHashErr(fileHash)
	if err != nil {
		return nil, err
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	_, isWriting := this.writingFileMap[fileHash]
	if isWriting {
		err = ErrFileIsWriting
		return
	}
	this.writingFileMap[fileHash] = zero.Zero{}

	err = this.checkStatus()
	if err != nil {
		return
	}

	return NewFileWriter(this, fileHash, bodySize, isPartial)
}

func (this *BlocksFile) OpenFileReader(fileHash string, isPartial bool) (*FileReader, error) {
	err := CheckHashErr(fileHash)
	if err != nil {
		return nil, err
	}

	this.mu.RLock()
	err = this.checkStatus()
	this.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	// 是否存在
	header, ok := this.mFile.CloneFileHeader(fileHash)
	if !ok {
		return nil, os.ErrNotExist
	}

	// TODO 对于partial content，需要传入ranges，用来判断是否有交集

	if header.IsWriting {
		return nil, ErrFileIsWriting
	}

	if !isPartial && !header.IsCompleted {
		return nil, os.ErrNotExist
	}

	// 先尝试从Pool中获取
	select {
	case reader := <-this.readerPool:
		if reader == nil {
			return nil, ErrClosed
		}
		reader.Reset(header)
		atomic.AddInt32(&this.countRefs, 1)
		return reader, nil
	default:
	}

	AckReadThread()
	fp, err := os.Open(this.fp.Name())
	ReleaseReadThread()
	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&this.countRefs, 1)
	return NewFileReader(this, fp, header), nil
}

func (this *BlocksFile) CloseFileReader(reader *FileReader) error {
	defer atomic.AddInt32(&this.countRefs, -1)

	select {
	case this.readerPool <- reader:
		return nil
	default:
		return reader.Free()
	}
}

func (this *BlocksFile) ExistFile(fileHash string) bool {
	err := CheckHashErr(fileHash)
	if err != nil {
		return false
	}

	return this.mFile.ExistFile(fileHash)
}

func (this *BlocksFile) RemoveFile(fileHash string) error {
	err := CheckHashErr(fileHash)
	if err != nil {
		return err
	}

	return this.mFile.RemoveFile(fileHash)
}

func (this *BlocksFile) Sync() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	err := this.checkStatus()
	if err != nil {
		return err
	}

	return this.sync(false)
}

func (this *BlocksFile) ForceSync() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	err := this.checkStatus()
	if err != nil {
		return err
	}

	return this.sync(true)
}

func (this *BlocksFile) SyncAt() time.Time {
	return this.syncAt
}

func (this *BlocksFile) Compact() error {
	// TODO 需要实现
	return nil
}

func (this *BlocksFile) RemoveAll() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.isClosed = true

	_ = this.mFile.RemoveAll()

	this.closeReaderPool()

	_ = this.fp.Close()
	return os.Remove(this.fp.Name())
}

// CanClose 检查是否可以关闭
func (this *BlocksFile) CanClose() bool {
	this.mu.RLock()
	defer this.mu.RUnlock()

	if len(this.writingFileMap) > 0 || atomic.LoadInt32(&this.countRefs) > 0 {
		return false
	}

	this.isClosing = true
	return true
}

// Close 关闭当前文件
func (this *BlocksFile) Close() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.isClosed {
		return nil
	}

	// TODO 决定是否同步
	//_ = this.sync(true)

	this.isClosed = true

	_ = this.mFile.Close()

	this.closeReaderPool()

	return this.fp.Close()
}

// IsClosing 判断当前文件是否正在关闭或者已关闭
func (this *BlocksFile) IsClosing() bool {
	return this.isClosed || this.isClosing
}

func (this *BlocksFile) IncrRef() {
	atomic.AddInt32(&this.countRefs, 1)
}

func (this *BlocksFile) DecrRef() {
	atomic.AddInt32(&this.countRefs, -1)
}

func (this *BlocksFile) TestReaderPool() chan *FileReader {
	return this.readerPool
}

func (this *BlocksFile) removeWritingFile(hash string) {
	this.mu.Lock()
	delete(this.writingFileMap, hash)
	this.mu.Unlock()
}

func (this *BlocksFile) checkStatus() error {
	if this.isClosed || this.isClosing {
		return fmt.Errorf("check status failed: %w", ErrClosed)
	}
	return nil
}

func (this *BlocksFile) currentPos() (int64, error) {
	return this.fp.Seek(0, io.SeekCurrent)
}

func (this *BlocksFile) sync(force bool) error {
	if !force {
		if this.writtenBytes < this.opt.BytesPerSync {
			return nil
		}
	}

	if this.writtenBytes > 0 {
		AckWriteThread()
		err := this.fp.Sync()
		ReleaseWriteThread()
		if err != nil {
			return err
		}
	}

	this.writtenBytes = 0

	this.syncAt = time.Now()

	if force {
		return this.mFile.SyncUnsafe()
	}

	return nil
}

func (this *BlocksFile) closeReaderPool() {
	for {
		select {
		case reader := <-this.readerPool:
			if reader != nil {
				_ = reader.Free()
			}
		default:
			return
		}
	}
}
