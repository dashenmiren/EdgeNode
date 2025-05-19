package nodes

import (
	"github.com/dashenmiren/EdgeCommon/pkg/nodeutils"
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"
	"github.com/dashenmiren/EdgeNode/internal/remotelogs"
)

// 健康检查
func (this *HTTPRequest) doHealthCheck(key string, isHealthCheck *bool) (stop bool) {
	this.tags = append(this.tags, "healthCheck")

	this.RawReq.Header.Del(serverconfigs.HealthCheckHeaderName)

	data, err := nodeutils.Base64DecodeMap(key)
	if err != nil {
		remotelogs.Error("HTTP_REQUEST_HEALTH_CHECK", "decode key failed: "+err.Error())
		return
	}
	*isHealthCheck = true

	this.web.StatRef = nil

	if !data.GetBool("accessLogIsOn") {
		this.disableLog = true
	}

	if data.GetBool("onlyBasicRequest") {
		return true
	}

	return
}
