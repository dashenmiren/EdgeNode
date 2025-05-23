// Copyright 2021 GoEdge goedge.cdn@gmail.com. All rights reserved.

package iplibrary_test

import (
	"github.com/dashenmiren/EdgeCommon/pkg/rpc/pb"
	"github.com/dashenmiren/EdgeNode/internal/iplibrary"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/logs"
	"testing"
	"time"
)

func TestSQLiteIPList_AddItem(t *testing.T) {
	db, err := iplibrary.NewSQLiteIPList()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	err = db.AddItem(&pb.IPItem{
		Id:                            1,
		IpFrom:                        "192.168.1.101",
		IpTo:                          "",
		Version:                       1024,
		ExpiredAt:                     time.Now().Unix() + 3600,
		Reason:                        "",
		ListId:                        2,
		IsDeleted:                     false,
		Type:                          "ipv4",
		EventLevel:                    "error",
		ListType:                      "black",
		IsGlobal:                      true,
		CreatedAt:                     0,
		NodeId:                        11,
		ServerId:                      22,
		SourceNodeId:                  0,
		SourceServerId:                0,
		SourceHTTPFirewallPolicyId:    0,
		SourceHTTPFirewallRuleGroupId: 0,
		SourceHTTPFirewallRuleSetId:   0,
		SourceServer:                  nil,
		SourceHTTPFirewallPolicy:      nil,
		SourceHTTPFirewallRuleGroup:   nil,
		SourceHTTPFirewallRuleSet:     nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("ok")
}

func TestSQLiteIPList_ReadItems(t *testing.T) {
	db, err := iplibrary.NewSQLiteIPList()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	defer func() {
		_ = db.Close()
	}()

	items, goNext, err := db.ReadItems(0, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("goNext:", goNext)
	logs.PrintAsJSON(items, t)
}

func TestSQLiteIPList_ReadMaxVersion(t *testing.T) {
	db, err := iplibrary.NewSQLiteIPList()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	t.Log(db.ReadMaxVersion())
}

func TestSQLiteIPList_UpdateMaxVersion(t *testing.T) {
	db, err := iplibrary.NewSQLiteIPList()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	err = db.UpdateMaxVersion(1027)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(db.ReadMaxVersion())
}
