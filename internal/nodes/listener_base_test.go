package nodes

import (
	"context"
	"testing"
	"time"

	"github.com/dashenmiren/EdgeCommon/pkg/nodeconfigs"
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/types"
)

func TestBaseListener_FindServer(t *testing.T) {
	sharedNodeConfig = &nodeconfigs.NodeConfig{}

	var listener = &BaseListener{}
	listener.Group = serverconfigs.NewServerAddressGroup("https://*:443")
	for i := 0; i < 1_000_000; i++ {
		var server = &serverconfigs.ServerConfig{
			IsOn: true,
			Name: types.String(i) + ".hello.com",
			ServerNames: []*serverconfigs.ServerNameConfig{
				{Name: types.String(i) + ".hello.com"},
			},
		}
		_ = server.Init(context.Background())
		listener.Group.Add(server)
	}

	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()

	t.Log(listener.findNamedServerMatched("855555.hello.com"))
}
