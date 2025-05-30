// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package bfs_test

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/bfs"
	"github.com/dashenmiren/EdgeNode/internal/utils/fasttime"
	"github.com/dashenmiren/EdgeNode/internal/utils/linkedlist"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"io"
	"os"
	"testing"
)

func TestFS_OpenFileWriter(t *testing.T) {
	fs, openErr := bfs.OpenFS(Tea.Root+"/data/bfs/test", bfs.DefaultFSOptions)
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer func() {
		_ = fs.Close()
	}()

	{
		writer, err := fs.OpenFileWriter(bfs.Hash("123456"), -1, false)
		if err != nil {
			t.Fatal(err)
		}

		err = writer.WriteMeta(200, fasttime.Now().Unix()+3600, -1)
		if err != nil {
			t.Fatal(err)
		}

		_, err = writer.WriteBody([]byte("Hello, World"))
		if err != nil {
			t.Fatal(err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		writer, err := fs.OpenFileWriter(bfs.Hash("654321"), 100, true)
		if err != nil {
			t.Fatal(err)
		}

		_, err = writer.WriteBody([]byte("Hello, World"))
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestFS_OpenFileReader(t *testing.T) {
	fs, openErr := bfs.OpenFS(Tea.Root+"/data/bfs/test", bfs.DefaultFSOptions)
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer func() {
		_ = fs.Close()
	}()

	reader, err := fs.OpenFileReader(bfs.Hash("123456"), false)
	if err != nil {
		if bfs.IsNotExist(err) {
			t.Log(err)
			return
		}
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
	logs.PrintAsJSON(reader.FileHeader(), t)
}

func TestFS_ExistFile(t *testing.T) {
	fs, openErr := bfs.OpenFS(Tea.Root+"/data/bfs/test", bfs.DefaultFSOptions)
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer func() {
		_ = fs.Close()
	}()

	exist, err := fs.ExistFile(bfs.Hash("123456"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("exist:", exist)
}

func TestFS_RemoveFile(t *testing.T) {
	fs, openErr := bfs.OpenFS(Tea.Root+"/data/bfs/test", bfs.DefaultFSOptions)
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer func() {
		_ = fs.Close()
	}()

	var hash = bfs.Hash("123456")
	err := fs.RemoveFile(hash)
	if err != nil {
		t.Fatal(err)
	}

	exist, err := fs.ExistFile(bfs.Hash("123456"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("exist:", exist)
}

func TestFS_OpenFileWriter_Close(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	fs, openErr := bfs.OpenFS(Tea.Root+"/data/bfs/test", &bfs.FSOptions{
		MaxOpenFiles: 99,
	})
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer func() {
		_ = fs.Close()
	}()

	var count = 2
	if testutils.IsSingleTesting() {
		count = 100
	}

	for i := 0; i < count; i++ {
		//t.Log("open", i)
		writer, err := fs.OpenFileWriter(bfs.Hash(types.String(i)), -1, false)
		if err != nil {
			t.Fatal(err)
		}
		_ = writer.Close()
	}

	t.Log(len(fs.TestBMap()), "block files, pid:", os.Getpid())

	var p = func() {
		var bNames []string
		fs.TestBList().Range(func(item *linkedlist.Item[string]) (goNext bool) {
			bNames = append(bNames, item.Value)
			return true
		})

		if len(bNames) != len(fs.TestBMap()) {
			t.Fatal("len(bNames)!=len(bMap)")
		}

		if len(bNames) < 10 {
			t.Log("["+types.String(len(bNames))+"]", bNames)
		} else {
			t.Log("["+types.String(len(bNames))+"]", bNames[:10], "...")
		}
	}

	p()

	{
		writer, err := fs.OpenFileWriter(bfs.Hash(types.String(10)), -1, false)
		if err != nil {
			t.Fatal(err)
		}
		_ = writer.Close()
	}

	p()

	// testing closing
	for i := 0; i < 3; i++ {
		writer, err := fs.OpenFileWriter(bfs.Hash(types.String(0)), -1, false)
		if err != nil {
			t.Fatal(err)
		}
		_ = writer.Close()
	}

	p()
}
