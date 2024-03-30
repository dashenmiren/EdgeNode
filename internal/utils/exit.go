package utils

import (
	"os"

	"github.com/TeaOSLab/EdgeNode/internal/events"
)

func Exit() {
	events.Notify(events.EventTerminated)
	os.Exit(0)
}
