package rpc_test

import (
	"sync"
	"testing"
	"time"

	"github.com/dashenmiren/EdgeCommon/pkg/rpc/pb"
	"github.com/dashenmiren/EdgeNode/internal/rpc"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	_ "github.com/iwind/TeaGo/bootstrap"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

func TestRPCConcurrentCall(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	rpcClient, err := rpc.SharedRPC()
	if err != nil {
		t.Fatal(err)
	}

	var before = time.Now()
	defer func() {
		t.Log("cost:", time.Since(before).Seconds()*1000, "ms")
	}()

	var concurrent = 3

	var wg = sync.WaitGroup{}
	wg.Add(concurrent)

	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()

			_, err = rpcClient.NodeRPC.FindCurrentNodeConfig(rpcClient.Context(), &pb.FindCurrentNodeConfigRequest{})
			if err != nil {
				t.Log(err)
			}
		}()
	}

	wg.Wait()
}

func TestRPC_Retry(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	rpcClient, err := rpc.SharedRPC()
	if err != nil {
		t.Fatal(err)
	}

	var ticker = time.NewTicker(1 * time.Second)
	for range ticker.C {
		go func() {
			_, err = rpcClient.NodeRPC.FindCurrentNodeConfig(rpcClient.Context(), &pb.FindCurrentNodeConfigRequest{})
			if err != nil {
				t.Log(timeutil.Format("H:i:s"), err)
			} else {
				t.Log(timeutil.Format("H:i:s"), "success")
			}
		}()
	}
}
