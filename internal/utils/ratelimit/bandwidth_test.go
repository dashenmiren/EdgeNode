package ratelimit_test

import (
	"context"
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/utils/ratelimit"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
)

func TestBandwidth(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var bandwidth = ratelimit.NewBandwidth(32 << 10)
	bandwidth.Ack(context.Background(), 123)
	bandwidth.Ack(context.Background(), 16<<10)
	bandwidth.Ack(context.Background(), 32<<10)
}

func TestBandwidth_0(t *testing.T) {
	var bandwidth = ratelimit.NewBandwidth(0)
	bandwidth.Ack(context.Background(), 123)
	bandwidth.Ack(context.Background(), 123456)
}
