package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, INFO)
	l.Info("test message", "key", "value")

	out := buf.String()
	if !strings.Contains(out, "level=INFO") {
		t.Errorf("expected level=INFO in output: %s", out)
	}
	if !strings.Contains(out, `msg="test message"`) {
		t.Errorf("expected msg in output: %s", out)
	}
	if !strings.Contains(out, "key=value") {
		t.Errorf("expected key=value in output: %s", out)
	}
}

func TestLogger_Debug_BelowLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, INFO) // DEBUG is below INFO — should be silenced
	l.Debug("should not appear")

	if buf.Len() > 0 {
		t.Errorf("expected no output for DEBUG below INFO level, got: %s", buf.String())
	}
}

func TestLogger_Discard(t *testing.T) {
	l := Discard()
	// must not panic
	l.Info("ignored")
	l.Warn("ignored")
	l.Error("ignored")
}

func TestLogger_MultipleFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, INFO)
	l.Info("request", "method", "GET", "path", "/", "status", 200)

	out := buf.String()
	for _, want := range []string{"method=GET", "path=/", "status=200"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output: %s", want, out)
		}
	}
}
