//go:build !plus

package nodes

import (
	"errors"
	"net/http"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

func (this *HTTPRequest) doOSSOrigin(origin *serverconfigs.OriginConfig) (resp *http.Response, goNext bool, errorCode string, ossBucketName string, err error) {
	// stub
	return nil, false, "", "", errors.New("not implemented")
}
