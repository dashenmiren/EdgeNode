// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package nodes_test

import (
	"github.com/dashenmiren/EdgeCommon/pkg/nodeconfigs"
	"github.com/dashenmiren/EdgeNode/internal/caches"
	"github.com/dashenmiren/EdgeNode/internal/nodes"
	"testing"
)

func TestHTTPCacheTaskManager_Loop(t *testing.T) {
	// initialize cache policies
	config, err := nodeconfigs.SharedNodeConfig()
	if err != nil {
		t.Fatal(err)
	}
	caches.SharedManager.UpdatePolicies(config.HTTPCachePolicies)

	var manager = nodes.NewHTTPCacheTaskManager()
	err = manager.Loop()
	if err != nil {
		t.Fatal(err)
	}
}
