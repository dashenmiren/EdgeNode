package waf

import "github.com/iwind/TeaGo/maps"

type ActionConfig struct {
	Code    string   `yaml:"code" json:"code"`
	Options maps.Map `yaml:"options" json:"options"`
}
