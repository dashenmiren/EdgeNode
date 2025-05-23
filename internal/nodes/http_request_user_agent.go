// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package nodes

import (
	"net/http"
)

func (this *HTTPRequest) doCheckUserAgent() (shouldStop bool) {
	if this.web.UserAgent == nil || !this.web.UserAgent.IsOn {
		return
	}

	const cacheSeconds = "3600" // 时间不能过长，防止修改设置后长期无法生效

	if this.web.UserAgent.MatchURL(this.URL()) && !this.web.UserAgent.AllowRequest(this.RawReq) {
		this.tags = append(this.tags, "userAgentCheck")
		this.writer.Header().Set("Cache-Control", "max-age="+cacheSeconds)
		this.writeCode(http.StatusForbidden, "The User-Agent has been blocked.", "当前访问已被UA名单拦截。")
		return true
	}

	return
}
