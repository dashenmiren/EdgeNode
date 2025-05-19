package waf

import (
	"github.com/dashenmiren/EdgeNode/internal/waf/requests"
	"net/http"
)

type AllowAction struct {
}

func (this *AllowAction) Perform(waf *WAF, request *requests.Request, writer http.ResponseWriter) (allow bool) {
	// do nothing
	return true
}
