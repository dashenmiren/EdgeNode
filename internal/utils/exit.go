// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved.

package utils

import (
	"os"

	"github.com/dashenmiren/EdgeNode/internal/events"
)

func Exit() {
	events.Notify(events.EventTerminated)
	os.Exit(0)
}
