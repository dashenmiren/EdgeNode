package configs_test

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/configs"
	_ "github.com/iwind/TeaGo/bootstrap"
	"gopkg.in/yaml.v3"
)

func TestLoadAPIConfig(t *testing.T) {
	config, err := configs.LoadAPIConfig()
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
