package io2

import (
	"context"
	"io"
)

// NewContextReader returns an io.Reader that returns ctx.Err() once ctx is
// canceled. The context is checked before each Read call, so an in-flight
// Read on the underlying reader is not interrupted — use this with readers
// that return control reasonably often (e.g. network or buffered readers).
func NewContextReader(ctx context.Context, r io.Reader) io.Reader {
	return &contextReader{ctx: ctx, r: r}
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *contextReader) Read(p []byte) (int, error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.r.Read(p)
}

// NewContextWriter returns an io.Writer that returns ctx.Err() once ctx is
// canceled. The context is checked before each Write call.
func NewContextWriter(ctx context.Context, w io.Writer) io.Writer {
	return &contextWriter{ctx: ctx, w: w}
}

type contextWriter struct {
	ctx context.Context
	w   io.Writer
}

func (cw *contextWriter) Write(p []byte) (int, error) {
	if err := cw.ctx.Err(); err != nil {
		return 0, err
	}
	return cw.w.Write(p)
}
