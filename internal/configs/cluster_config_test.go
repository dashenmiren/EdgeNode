// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package configs_test

import (
	"github.com/dashenmiren/EdgeNode/internal/configs"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestLoadClusterConfig(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	config, err := configs.LoadClusterConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", config)

	configData, err := yaml.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(configData))
}
