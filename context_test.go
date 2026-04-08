package io2

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestContextReader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := NewContextReader(ctx, strings.NewReader("abcdef"))

	buf := make([]byte, 3)
	n, err := r.Read(buf)
	if err != nil || n != 3 || string(buf) != "abc" {
		t.Fatalf("Read before cancel: n=%d buf=%q err=%v", n, buf, err)
	}

	cancel()
	_, err = r.Read(buf)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Read after cancel err = %v; want context.Canceled", err)
	}
}

func TestContextWriter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var buf bytes.Buffer
	w := NewContextWriter(ctx, &buf)

	if _, err := w.Write([]byte("abc")); err != nil {
		t.Fatalf("Write before cancel: %v", err)
	}

	cancel()
	_, err := w.Write([]byte("def"))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Write after cancel err = %v; want context.Canceled", err)
	}
	if buf.String() != "abc" {
		t.Errorf("buf = %q; want %q", buf.String(), "abc")
	}
}

func TestContextReaderEOF(t *testing.T) {
	r := NewContextReader(context.Background(), strings.NewReader(""))
	_, err := r.Read(make([]byte, 4))
	if !errors.Is(err, io.EOF) {
		t.Errorf("err = %v; want io.EOF", err)
	}
}
