package iplibrary

import (
	"github.com/dashenmiren/EdgeCommon/pkg/rpc/pb"
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"testing"
)

func TestHTTPAPIAction_AddItem(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var action = NewHTTPAPIAction()
	action.config = &firewallconfigs.FirewallActionHTTPAPIConfig{
		URL:            "http://127.0.0.1:2345/post",
		TimeoutSeconds: 0,
	}
	err := action.AddItem(IPListTypeBlack, &pb.IPItem{
		Type:   "ipv4",
		Id:     1,
		IpFrom: "192.168.1.100",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestHTTPAPIAction_DeleteItem(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var action = NewHTTPAPIAction()
	action.config = &firewallconfigs.FirewallActionHTTPAPIConfig{
		URL:            "http://127.0.0.1:2345/post",
		TimeoutSeconds: 0,
	}
	err := action.DeleteItem(IPListTypeBlack, &pb.IPItem{
		Type:   "ipv4",
		Id:     1,
		IpFrom: "192.168.1.100",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
