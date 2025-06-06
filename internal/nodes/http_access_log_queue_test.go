// Copyright 2021 GoEdge goedge.cdn@gmail.com. All rights reserved.

package nodes_test

import (
	"bytes"
	"github.com/dashenmiren/EdgeCommon/pkg/rpc/pb"
	"github.com/dashenmiren/EdgeNode/internal/nodes"
	"github.com/dashenmiren/EdgeNode/internal/rpc"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	_ "github.com/iwind/TeaGo/bootstrap"
	"google.golang.org/grpc/status"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestHTTPAccessLogQueue_Push(t *testing.T) {
	// 发送到API
	client, err := rpc.SharedRPC()
	if err != nil {
		t.Fatal(err)
	}

	var requestId = 1_000_000

	var utf8Bytes = []byte{}
	for i := 0; i < 254; i++ {
		utf8Bytes = append(utf8Bytes, uint8(i))
	}

	//bytes = []byte("真不错")

	var accessLog = &pb.HTTPAccessLog{
		ServerId:    23,
		RequestId:   strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(requestId) + strconv.FormatInt(1, 10),
		NodeId:      48,
		Host:        "www.hello.com",
		RequestURI:  string(utf8Bytes),
		RequestPath: string(utf8Bytes),
		Timestamp:   time.Now().Unix(),
		Cookie:      map[string]string{"test": string(utf8Bytes)},

		Header: map[string]*pb.Strings{
			"test": {Values: []string{string(utf8Bytes)}},
		},
	}

	new(nodes.HTTPAccessLogQueue).ToValidUTF8(accessLog)

	//	logs.PrintAsJSON(accessLog)

	//t.Log(strings.ToValidUTF8(string(utf8Bytes), ""))
	_, err = client.HTTPAccessLogRPC.CreateHTTPAccessLogs(client.Context(), &pb.CreateHTTPAccessLogsRequest{HttpAccessLogs: []*pb.HTTPAccessLog{
		accessLog,
	}})
	if err != nil {
		// 这里只是为了重现错误
		t.Logf("%#v, %s", err, err.Error())

		statusErr, ok := status.FromError(err)
		if ok {
			t.Logf("%#v", statusErr)
		}
		return
	}
	t.Log("ok")
}

func TestHTTPAccessLogQueue_Push2(t *testing.T) {
	var utf8Bytes = []byte{}
	for i := 0; i < 254; i++ {
		utf8Bytes = append(utf8Bytes, uint8(i))
	}

	var accessLog = &pb.HTTPAccessLog{
		ServerId:    23,
		RequestId:   strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(1) + strconv.FormatInt(1, 10),
		NodeId:      48,
		Host:        "www.hello.com",
		RequestURI:  string(utf8Bytes),
		RequestPath: string(utf8Bytes),
		Timestamp:   time.Now().Unix(),
	}
	var v = reflect.Indirect(reflect.ValueOf(accessLog))
	var countFields = v.NumField()
	for i := 0; i < countFields; i++ {
		var field = v.Field(i)
		if field.Kind() == reflect.String {
			field.SetString(strings.ToValidUTF8(field.String(), ""))
		}
	}

	client, err := rpc.SharedRPC()
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.HTTPAccessLogRPC.CreateHTTPAccessLogs(client.Context(), &pb.CreateHTTPAccessLogsRequest{HttpAccessLogs: []*pb.HTTPAccessLog{
		accessLog,
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestHTTPAccessLogQueue_Memory(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	testutils.StartMemoryStats(t)

	debug.SetGCPercent(10)

	var accessLogs = []*pb.HTTPAccessLog{}
	for i := 0; i < 20_000; i++ {
		accessLogs = append(accessLogs, &pb.HTTPAccessLog{
			RequestPath: "https://cdn.foyeseo.com/hello/world",
		})
	}

	runtime.GC()
	_ = accessLogs

	// will not release automatically
	func() {
		var accessLogs1 = []*pb.HTTPAccessLog{}
		for i := 0; i < 2_000_000; i++ {
			accessLogs1 = append(accessLogs1, &pb.HTTPAccessLog{
				RequestPath: "https://cdn.foyeseo.com/hello/world",
			})
		}
		_ = accessLogs1
	}()

	time.Sleep(5 * time.Second)
}

func TestUTF8_IsValid(t *testing.T) {
	t.Log(utf8.ValidString("abc"))

	var noneUTF8Bytes = []byte{}
	for i := 0; i < 254; i++ {
		noneUTF8Bytes = append(noneUTF8Bytes, uint8(i))
	}
	t.Log(utf8.ValidString(string(noneUTF8Bytes)))
}

func BenchmarkHTTPAccessLogQueue_ToValidUTF8(b *testing.B) {
	runtime.GOMAXPROCS(1)

	var utf8Bytes = []byte{}
	for i := 0; i < 254; i++ {
		utf8Bytes = append(utf8Bytes, uint8(i))
	}

	for i := 0; i < b.N; i++ {
		_ = bytes.ToValidUTF8(utf8Bytes, nil)
	}
}

func BenchmarkHTTPAccessLogQueue_ToValidUTF8String(b *testing.B) {
	runtime.GOMAXPROCS(1)

	var utf8Bytes = []byte{}
	for i := 0; i < 254; i++ {
		utf8Bytes = append(utf8Bytes, uint8(i))
	}

	var s = string(utf8Bytes)
	for i := 0; i < b.N; i++ {
		_ = strings.ToValidUTF8(s, "")
	}
}

func BenchmarkAppendAccessLogs(b *testing.B) {
	b.ReportAllocs()

	var stat1 = &runtime.MemStats{}
	runtime.ReadMemStats(stat1)

	const count = 20000
	var a = make([]*pb.HTTPAccessLog, 0, count)
	for i := 0; i < b.N; i++ {
		a = append(a, &pb.HTTPAccessLog{
			RequestPath: "/hello/world",
			Host:        "example.com",
			RequestBody: bytes.Repeat([]byte{'A'}, 1024),
		})
		if len(a) == count {
			a = make([]*pb.HTTPAccessLog, 0, count)
		}
	}

	_ = len(a)

	var stat2 = &runtime.MemStats{}
	runtime.ReadMemStats(stat2)
	b.Log((stat2.TotalAlloc-stat1.TotalAlloc)>>20, "MB")
}

func BenchmarkAppendAccessLogs2(b *testing.B) {
	b.ReportAllocs()

	var stat1 = &runtime.MemStats{}
	runtime.ReadMemStats(stat1)

	const count = 20000
	var a = []*pb.HTTPAccessLog{}
	for i := 0; i < b.N; i++ {
		a = append(a, &pb.HTTPAccessLog{
			RequestPath: "/hello/world",
			Host:        "example.com",
			RequestBody: bytes.Repeat([]byte{'A'}, 1024),
		})
		if len(a) == count {
			a = []*pb.HTTPAccessLog{}
		}
	}

	_ = len(a)

	var stat2 = &runtime.MemStats{}
	runtime.ReadMemStats(stat2)
	b.Log((stat2.TotalAlloc-stat1.TotalAlloc)>>20, "MB")
}
