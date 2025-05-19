package stats

import (
	"github.com/mssola/useragent"
)

type UserAgentParserResult struct {
	OS             useragent.OSInfo
	BrowserName    string
	BrowserVersion string
	IsMobile       bool
}
