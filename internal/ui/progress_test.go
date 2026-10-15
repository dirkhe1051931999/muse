package ui

import (
	"bytes"
	"testing"
	"time"
)

func TestStartGeneratingStopsCleanly(t *testing.T) {
	var buf bytes.Buffer
	stop := StartGenerating(&buf)
	time.Sleep(150 * time.Millisecond)
	stop()
	stop() // idempotent
	if buf.Len() == 0 {
		t.Fatalf("expected progress output")
	}
}
