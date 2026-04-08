package io2

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestCountingReader(t *testing.T) {
	cr := NewCountingReader(strings.NewReader("hello world"))
	buf, err := io.ReadAll(cr)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(buf) != "hello world" {
		t.Errorf("bytes = %q; want %q", buf, "hello world")
	}
	if got := cr.N(); got != 11 {
		t.Errorf("N() = %d; want 11", got)
	}
}

func TestCountingWriter(t *testing.T) {
	var buf bytes.Buffer
	cw := NewCountingWriter(&buf)
	if _, err := cw.Write([]byte("hello")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := cw.Write([]byte(" world")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.String() != "hello world" {
		t.Errorf("buf = %q; want %q", buf.String(), "hello world")
	}
	if got := cw.N(); got != 11 {
		t.Errorf("N() = %d; want 11", got)
	}
}
