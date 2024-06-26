package checkpoints

import (
	"github.com/dashenmiren/EdgeNode/internal/waf/requests"
	"github.com/dashenmiren/EdgeNode/internal/waf/utils"
	"github.com/iwind/TeaGo/maps"
)

// RequestAllCheckpoint ${requestAll}
type RequestAllCheckpoint struct {
	Checkpoint
}

func (this *RequestAllCheckpoint) RequestValue(req requests.Request, param string, options maps.Map, ruleId int64) (value any, hasRequestBody bool, sysErr error, userErr error) {
	var valueBytes = [][]byte{}
	if len(req.WAFRaw().RequestURI) > 0 {
		valueBytes = append(valueBytes, []byte(req.WAFRaw().RequestURI))
	} else if req.WAFRaw().URL != nil {
		valueBytes = append(valueBytes, []byte(req.WAFRaw().URL.RequestURI()))
	}

	if this.RequestBodyIsEmpty(req) {
		value = valueBytes
		return
	}

	if req.WAFRaw().Body != nil {
		var bodyData = req.WAFGetCacheBody()
		hasRequestBody = true
		if len(bodyData) == 0 {
			data, err := req.WAFReadBody(req.WAFMaxRequestSize()) // read body
			if err != nil {
				return "", hasRequestBody, err, nil
			}

			bodyData = data
			req.WAFSetCacheBody(data)
			req.WAFRestoreBody(data)
		}
		if len(bodyData) > 0 {
			valueBytes = append(valueBytes, bodyData)
		}
	}

	value = valueBytes

	return
}

func (this *RequestAllCheckpoint) ResponseValue(req requests.Request, resp *requests.Response, param string, options maps.Map, ruleId int64) (value any, hasRequestBody bool, sysErr error, userErr error) {
	value = ""
	if this.IsRequest() {
		return this.RequestValue(req, param, options, ruleId)
	}
	return
}

func (this *RequestAllCheckpoint) CacheLife() utils.CacheLife {
	return utils.CacheShortLife
}
