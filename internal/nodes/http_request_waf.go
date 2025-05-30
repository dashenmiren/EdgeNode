package nodes

import (
	"bytes"
	iplib "github.com/dashenmiren/EdgeCommon/pkg/iplibrary"
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/dashenmiren/EdgeNode/internal/iplibrary"
	"github.com/dashenmiren/EdgeNode/internal/remotelogs"
	"github.com/dashenmiren/EdgeNode/internal/stats"
	"github.com/dashenmiren/EdgeNode/internal/waf"
	wafutils "github.com/dashenmiren/EdgeNode/internal/waf/utils"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/types"
	"io"
	"net/http"
	"time"
)

// 调用WAF
func (this *HTTPRequest) doWAFRequest() (blocked bool) {
	if this.web.FirewallRef == nil || !this.web.FirewallRef.IsOn {
		return
	}

	var remoteAddr = this.requestRemoteAddr(true)

	// 检查是否为白名单直连
	if !Tea.IsTesting() && this.nodeConfig.IPIsAutoAllowed(remoteAddr) {
		return
	}

	// 当前连接是否已关闭
	if this.isConnClosed() {
		this.disableLog = true
		return true
	}

	// 是否在全局名单中
	canGoNext, isInAllowedList, _ := iplibrary.AllowIP(remoteAddr, this.ReqServer.Id)
	if !canGoNext {
		this.disableLog = true
		this.Close()
		return true
	}
	if isInAllowedList {
		return false
	}

	// 检查是否在临时黑名单中
	if waf.SharedIPBlackList.Contains(waf.IPTypeAll, firewallconfigs.FirewallScopeServer, this.ReqServer.Id, remoteAddr) || waf.SharedIPBlackList.Contains(waf.IPTypeAll, firewallconfigs.FirewallScopeGlobal, 0, remoteAddr) {
		this.disableLog = true
		this.Close()

		return true
	}

	var forceLog = false
	var forceLogRequestBody = false
	var forceLogRegionDenying = false
	if this.ReqServer.HTTPFirewallPolicy != nil &&
		this.ReqServer.HTTPFirewallPolicy.IsOn &&
		this.ReqServer.HTTPFirewallPolicy.Log != nil &&
		this.ReqServer.HTTPFirewallPolicy.Log.IsOn {
		forceLog = true
		forceLogRequestBody = this.ReqServer.HTTPFirewallPolicy.Log.RequestBody
		forceLogRegionDenying = this.ReqServer.HTTPFirewallPolicy.Log.RegionDenying
	}

	// 检查IP名单
	{
		// 当前服务的独立设置
		if this.web.FirewallPolicy != nil && this.web.FirewallPolicy.IsOn {
			blockedRequest, breakChecking := this.checkWAFRemoteAddr(this.web.FirewallPolicy)
			if blockedRequest {
				return true
			}
			if breakChecking {
				return false
			}
		}

		// 公用的防火墙设置
		if this.ReqServer.HTTPFirewallPolicy != nil && this.ReqServer.HTTPFirewallPolicy.IsOn {
			blockedRequest, breakChecking := this.checkWAFRemoteAddr(this.ReqServer.HTTPFirewallPolicy)
			if blockedRequest {
				return true
			}
			if breakChecking {
				return false
			}
		}
	}

	// 检查WAF规则
	{
		// 当前服务的独立设置
		if this.web.FirewallPolicy != nil && this.web.FirewallPolicy.IsOn {
			blockedRequest, breakChecking := this.checkWAFRequest(this.web.FirewallPolicy, forceLog, forceLogRequestBody, forceLogRegionDenying, false)
			if blockedRequest {
				return true
			}
			if breakChecking {
				return false
			}
		}

		// 公用的防火墙设置
		if this.ReqServer.HTTPFirewallPolicy != nil && this.ReqServer.HTTPFirewallPolicy.IsOn {
			blockedRequest, breakChecking := this.checkWAFRequest(this.ReqServer.HTTPFirewallPolicy, forceLog, forceLogRequestBody, forceLogRegionDenying, this.web.FirewallRef.IgnoreGlobalRules)
			if blockedRequest {
				return true
			}
			if breakChecking {
				return false
			}
		}
	}

	return
}

// check client remote address
func (this *HTTPRequest) checkWAFRemoteAddr(firewallPolicy *firewallconfigs.HTTPFirewallPolicy) (blocked bool, breakChecking bool) {
	if firewallPolicy == nil {
		return
	}

	var isDefendMode = firewallPolicy.Mode == firewallconfigs.FirewallModeDefend

	// 检查IP白名单
	var remoteAddrs []string
	if len(this.remoteAddr) > 0 {
		remoteAddrs = []string{this.remoteAddr}
	} else {
		remoteAddrs = this.requestRemoteAddrs()
	}

	var inbound = firewallPolicy.Inbound
	if inbound == nil {
		return
	}
	for _, ref := range inbound.AllAllowListRefs() {
		if ref.IsOn && ref.ListId > 0 {
			var list = iplibrary.SharedIPListManager.FindList(ref.ListId)
			if list != nil {
				_, found := list.ContainsIPStrings(remoteAddrs)
				if found {
					breakChecking = true
					return
				}
			}
		}
	}

	// 检查IP黑名单
	if isDefendMode {
		for _, ref := range inbound.AllDenyListRefs() {
			if ref.IsOn && ref.ListId > 0 {
				var list = iplibrary.SharedIPListManager.FindList(ref.ListId)
				if list != nil {
					item, found := list.ContainsIPStrings(remoteAddrs)
					if found {
						// 触发事件
						if item != nil && len(item.EventLevel) > 0 {
							actions := iplibrary.SharedActionManager.FindEventActions(item.EventLevel)
							for _, action := range actions {
								goNext, err := action.DoHTTP(this.RawReq, this.RawWriter)
								if err != nil {
									remotelogs.Error("HTTP_REQUEST_WAF", "do action '"+err.Error()+"' failed: "+err.Error())
									return true, false
								}
								if !goNext {
									this.disableLog = true
									return true, false
								}
							}
						}

						// TODO 考虑是否需要记录日志信息吗，可能数据量非常庞大，所以暂时不记录

						this.writer.WriteHeader(http.StatusForbidden)
						this.writer.Close()

						// 停止日志
						this.disableLog = true

						return true, false
					}
				}
			}
		}
	}

	return
}

// check waf inbound rules
func (this *HTTPRequest) checkWAFRequest(firewallPolicy *firewallconfigs.HTTPFirewallPolicy, forceLog bool, logRequestBody bool, logDenying bool, ignoreRules bool) (blocked bool, breakChecking bool) {
	// 检查配置是否为空
	if firewallPolicy == nil || !firewallPolicy.IsOn || firewallPolicy.Inbound == nil || !firewallPolicy.Inbound.IsOn || firewallPolicy.Mode == firewallconfigs.FirewallModeBypass {
		return
	}

	var isDefendMode = firewallPolicy.Mode == firewallconfigs.FirewallModeDefend

	// 检查IP白名单
	var remoteAddrs []string
	if len(this.remoteAddr) > 0 {
		remoteAddrs = []string{this.remoteAddr}
	} else {
		remoteAddrs = this.requestRemoteAddrs()
	}

	var inbound = firewallPolicy.Inbound
	if inbound == nil {
		return
	}

	// 检查地区封禁
	if firewallPolicy.Inbound.Region != nil && firewallPolicy.Inbound.Region.IsOn {
		var regionConfig = firewallPolicy.Inbound.Region
		if regionConfig.IsNotEmpty() {
			for _, remoteAddr := range remoteAddrs {
				var result = iplib.LookupIP(remoteAddr)
				if result != nil && result.IsOk() {
					var currentURL = this.URL()
					if regionConfig.MatchCountryURL(currentURL) {
						// 检查国家/地区级别封禁
						if !regionConfig.IsAllowedCountry(result.CountryId(), result.ProvinceId()) && (!regionConfig.AllowSearchEngine || wafutils.CheckSearchEngine(remoteAddr)) {
							this.firewallPolicyId = firewallPolicy.Id

							if isDefendMode {
								var promptHTML string
								if len(regionConfig.CountryHTML) > 0 {
									promptHTML = regionConfig.CountryHTML
								} else if this.ReqServer != nil && this.ReqServer.HTTPFirewallPolicy != nil && len(this.ReqServer.HTTPFirewallPolicy.DenyCountryHTML) > 0 {
									promptHTML = this.ReqServer.HTTPFirewallPolicy.DenyCountryHTML
								}

								if len(promptHTML) > 0 {
									var formattedHTML = this.Format(promptHTML)
									this.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
									this.writer.Header().Set("Content-Length", types.String(len(formattedHTML)))
									this.writer.WriteHeader(http.StatusForbidden)
									_, _ = this.writer.Write([]byte(formattedHTML))
								} else {
									this.writeCode(http.StatusForbidden, "The region has been denied.", "当前区域禁止访问")
								}

								// 延时返回，避免攻击
								time.Sleep(1 * time.Second)
							}

							// 停止日志
							if !logDenying {
								this.disableLog = true
							} else {
								this.tags = append(this.tags, "denyCountry")
							}

							if isDefendMode {
								return true, false
							}
						}
					}

					if regionConfig.MatchProvinceURL(currentURL) {
						// 检查省份封禁
						if !regionConfig.IsAllowedProvince(result.CountryId(), result.ProvinceId()) {
							this.firewallPolicyId = firewallPolicy.Id

							if isDefendMode {
								var promptHTML string
								if len(regionConfig.ProvinceHTML) > 0 {
									promptHTML = regionConfig.ProvinceHTML
								} else if this.ReqServer != nil && this.ReqServer.HTTPFirewallPolicy != nil && len(this.ReqServer.HTTPFirewallPolicy.DenyProvinceHTML) > 0 {
									promptHTML = this.ReqServer.HTTPFirewallPolicy.DenyProvinceHTML
								}

								if len(promptHTML) > 0 {
									var formattedHTML = this.Format(promptHTML)
									this.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
									this.writer.Header().Set("Content-Length", types.String(len(formattedHTML)))
									this.writer.WriteHeader(http.StatusForbidden)
									_, _ = this.writer.Write([]byte(formattedHTML))
								} else {
									this.writeCode(http.StatusForbidden, "The region has been denied.", "当前区域禁止访问")
								}

								// 延时返回，避免攻击
								time.Sleep(1 * time.Second)
							}

							// 停止日志
							if !logDenying {
								this.disableLog = true
							} else {
								this.tags = append(this.tags, "denyProvince")
							}

							if isDefendMode {
								return true, false
							}
						}
					}
				}
			}
		}
	}

	// 是否执行规则
	if ignoreRules {
		return
	}

	// 规则测试
	var w = waf.SharedWAFManager.FindWAF(firewallPolicy.Id)
	if w == nil {
		return
	}

	result, err := w.MatchRequest(this, this.writer, this.web.FirewallRef.DefaultCaptchaType)
	if err != nil {
		if !this.canIgnore(err) {
			remotelogs.Warn("HTTP_REQUEST_WAF", this.rawURI+": "+err.Error())
		}
		return
	}
	if result.IsAllowed && (len(result.AllowScope) == 0 || result.AllowScope == waf.AllowScopeGlobal) {
		breakChecking = true
	}
	if forceLog && logRequestBody && result.HasRequestBody && result.Set != nil && result.Set.HasAttackActions() {
		this.wafHasRequestBody = true
	}

	if result.Set != nil {
		if forceLog {
			this.forceLog = true
		}

		if result.Set.HasSpecialActions() {
			this.firewallPolicyId = firewallPolicy.Id
			this.firewallRuleGroupId = types.Int64(result.Group.Id)
			this.firewallRuleSetId = types.Int64(result.Set.Id)

			if result.Set.HasAttackActions() {
				this.isAttack = true
			}

			// 添加统计
			stats.SharedHTTPRequestStatManager.AddFirewallRuleGroupId(this.ReqServer.Id, this.firewallRuleGroupId, result.Set.Actions)
		}

		this.firewallActions = append(result.Set.ActionCodes(), firewallPolicy.Mode)
	}

	return !result.GoNext, breakChecking
}

// call response waf
func (this *HTTPRequest) doWAFResponse(resp *http.Response) (blocked bool) {
	if this.web.FirewallRef == nil || !this.web.FirewallRef.IsOn {
		return
	}

	// 当前服务的独立设置
	var forceLog = false
	var forceLogRequestBody = false
	if this.ReqServer.HTTPFirewallPolicy != nil && this.ReqServer.HTTPFirewallPolicy.IsOn && this.ReqServer.HTTPFirewallPolicy.Log != nil && this.ReqServer.HTTPFirewallPolicy.Log.IsOn {
		forceLog = true
		forceLogRequestBody = this.ReqServer.HTTPFirewallPolicy.Log.RequestBody
	}

	if this.web.FirewallPolicy != nil && this.web.FirewallPolicy.IsOn {
		blockedRequest, breakChecking := this.checkWAFResponse(this.web.FirewallPolicy, resp, forceLog, forceLogRequestBody, false)
		if blockedRequest {
			return true
		}
		if breakChecking {
			return
		}
	}

	// 公用的防火墙设置
	if this.ReqServer.HTTPFirewallPolicy != nil && this.ReqServer.HTTPFirewallPolicy.IsOn {
		blockedRequest, _ := this.checkWAFResponse(this.ReqServer.HTTPFirewallPolicy, resp, forceLog, forceLogRequestBody, this.web.FirewallRef.IgnoreGlobalRules)
		if blockedRequest {
			return true
		}
	}
	return
}

// check waf outbound rules
func (this *HTTPRequest) checkWAFResponse(firewallPolicy *firewallconfigs.HTTPFirewallPolicy, resp *http.Response, forceLog bool, logRequestBody bool, ignoreRules bool) (blocked bool, breakChecking bool) {
	if firewallPolicy == nil || !firewallPolicy.IsOn || !firewallPolicy.Outbound.IsOn || firewallPolicy.Mode == firewallconfigs.FirewallModeBypass {
		return
	}

	// 是否执行规则
	if ignoreRules {
		return
	}

	var w = waf.SharedWAFManager.FindWAF(firewallPolicy.Id)
	if w == nil {
		return
	}

	result, err := w.MatchResponse(this, resp, this.writer)
	if err != nil {
		if !this.canIgnore(err) {
			remotelogs.Warn("HTTP_REQUEST_WAF", this.rawURI+": "+err.Error())
		}
		return
	}
	if result.IsAllowed && (len(result.AllowScope) == 0 || result.AllowScope == waf.AllowScopeGlobal) {
		breakChecking = true
	}
	if forceLog && logRequestBody && result.HasRequestBody && result.Set != nil && result.Set.HasAttackActions() {
		this.wafHasRequestBody = true
	}

	if result.Set != nil {
		if forceLog {
			this.forceLog = true
		}

		if result.Set.HasSpecialActions() {
			this.firewallPolicyId = firewallPolicy.Id
			this.firewallRuleGroupId = types.Int64(result.Group.Id)
			this.firewallRuleSetId = types.Int64(result.Set.Id)

			if result.Set.HasAttackActions() {
				this.isAttack = true
			}

			// 添加统计
			stats.SharedHTTPRequestStatManager.AddFirewallRuleGroupId(this.ReqServer.Id, this.firewallRuleGroupId, result.Set.Actions)
		}

		this.firewallActions = append(result.Set.ActionCodes(), firewallPolicy.Mode)
	}

	return !result.GoNext, breakChecking
}

// WAFRaw 原始请求
func (this *HTTPRequest) WAFRaw() *http.Request {
	return this.RawReq
}

// WAFRemoteIP 客户端IP
func (this *HTTPRequest) WAFRemoteIP() string {
	return this.requestRemoteAddr(true)
}

// WAFGetCacheBody 获取缓存中的Body
func (this *HTTPRequest) WAFGetCacheBody() []byte {
	return this.requestBodyData
}

// WAFSetCacheBody 设置Body
func (this *HTTPRequest) WAFSetCacheBody(body []byte) {
	this.requestBodyData = body
}

// WAFReadBody 读取Body
func (this *HTTPRequest) WAFReadBody(max int64) (data []byte, err error) {
	if this.RawReq.ContentLength > 0 {
		data, err = io.ReadAll(io.LimitReader(this.RawReq.Body, max))
	}

	return
}

// WAFRestoreBody 恢复Body
func (this *HTTPRequest) WAFRestoreBody(data []byte) {
	if len(data) > 0 {
		this.RawReq.Body = io.NopCloser(io.MultiReader(bytes.NewBuffer(data), this.RawReq.Body))
	}
}

// WAFServerId 服务ID
func (this *HTTPRequest) WAFServerId() int64 {
	return this.ReqServer.Id
}

// WAFClose 关闭连接
func (this *HTTPRequest) WAFClose() {
	this.Close()

	// 这里不要强关IP所有连接，避免因为单个服务而影响所有
}

func (this *HTTPRequest) WAFOnAction(action interface{}) (goNext bool) {
	if action == nil {
		return true
	}

	instance, ok := action.(waf.ActionInterface)
	if !ok {
		return true
	}

	switch instance.Code() {
	case waf.ActionTag:
		this.tags = append(this.tags, action.(*waf.TagAction).Tags...)
	}
	return true
}

func (this *HTTPRequest) WAFFingerprint() []byte {
	// 目前只有HTTPS请求才有指纹
	if !this.IsHTTPS {
		return nil
	}

	var requestConn = this.RawReq.Context().Value(HTTPConnContextKey)
	if requestConn == nil {
		return nil
	}

	clientConn, ok := requestConn.(ClientConnInterface)
	if ok {
		return clientConn.Fingerprint()
	}

	return nil
}

func (this *HTTPRequest) WAFMaxRequestSize() int64 {
	var maxRequestSize = firewallconfigs.DefaultMaxRequestBodySize
	if this.ReqServer.HTTPFirewallPolicy != nil && this.ReqServer.HTTPFirewallPolicy.MaxRequestBodySize > 0 {
		maxRequestSize = this.ReqServer.HTTPFirewallPolicy.MaxRequestBodySize
	}
	return maxRequestSize
}

// DisableAccessLog 在当前请求中不使用访问日志
func (this *HTTPRequest) DisableAccessLog() {
	this.disableLog = true
}

// DisableStat 停用统计
func (this *HTTPRequest) DisableStat() {
	if this.web != nil {
		this.web.StatRef = nil
	}

	this.disableMetrics = true
}
