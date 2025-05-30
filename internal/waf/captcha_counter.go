// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package waf

import (
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/dashenmiren/EdgeNode/internal/utils/counters"
	"github.com/dashenmiren/EdgeNode/internal/waf/requests"
	"github.com/iwind/TeaGo/types"
	"time"
)

type CaptchaPageCode = string

const (
	CaptchaPageCodeInit   CaptchaPageCode = "init"
	CaptchaPageCodeShow   CaptchaPageCode = "show"
	CaptchaPageCodeImage  CaptchaPageCode = "image"
	CaptchaPageCodeSubmit CaptchaPageCode = "submit"
)

// CaptchaIncreaseFails 增加Captcha失败次数，以便后续操作
func CaptchaIncreaseFails(req requests.Request, actionConfig *CaptchaAction, policyId int64, groupId int64, setId int64, pageCode CaptchaPageCode, useLocalFirewall bool) (goNext bool) {
	var maxFails = actionConfig.MaxFails
	var failBlockTimeout = actionConfig.FailBlockTimeout
	if maxFails > 0 && failBlockTimeout > 0 {
		if maxFails <= 3 {
			maxFails = 3 // 不能小于3，防止意外刷新出现
		}
		var countFails = counters.SharedCounter.IncreaseKey(CaptchaCacheKey(req, pageCode), 300)
		if int(countFails) >= maxFails {
			SharedIPBlackList.RecordIP(IPTypeAll, firewallconfigs.FirewallScopeServer, req.WAFServerId(), req.WAFRemoteIP(), time.Now().Unix()+int64(failBlockTimeout), policyId, useLocalFirewall, groupId, setId, "CAPTCHA验证连续失败超过"+types.String(maxFails)+"次")
			return false
		}
	}
	return true
}

// CaptchaDeleteCacheKey 清除计数
func CaptchaDeleteCacheKey(req requests.Request) {
	counters.SharedCounter.ResetKey(CaptchaCacheKey(req, CaptchaPageCodeInit))
	counters.SharedCounter.ResetKey(CaptchaCacheKey(req, CaptchaPageCodeShow))
	counters.SharedCounter.ResetKey(CaptchaCacheKey(req, CaptchaPageCodeImage))
	counters.SharedCounter.ResetKey(CaptchaCacheKey(req, CaptchaPageCodeSubmit))
}

// CaptchaCacheKey 获取Captcha缓存Key
func CaptchaCacheKey(req requests.Request, pageCode CaptchaPageCode) string {
	return "WAF:CAPTCHA:FAILS:" + pageCode + ":" + req.WAFRemoteIP() + ":" + types.String(req.WAFServerId())
}
