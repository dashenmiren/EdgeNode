package checkpoints_test

import (
	"net/http"
	"testing"

	"github.com/TeaOSLab/EdgeNode/internal/waf/checkpoints"
	"github.com/TeaOSLab/EdgeNode/internal/waf/requests"
)

func TestRequestHeaderNamesCheckpoint_RequestValue(t *testing.T) {
	var checkpoint = &checkpoints.RequestHeaderNamesCheckpoint{}
	rawReq, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	rawReq.Header.Set("Accept", "text/html")
	rawReq.Header.Set("User-Agent", "Chrome")
	rawReq.Header.Set("Accept-Encoding", "br, gzip")
	var req = requests.NewTestRequest(rawReq)
	t.Log(checkpoint.RequestValue(req, "", nil, 0))
}
