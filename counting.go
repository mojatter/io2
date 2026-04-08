package io2

import (
	"io"
	"sync/atomic"
)

// CountingReader wraps an io.Reader and counts the total number of bytes
// successfully read. It is safe for concurrent use.
type CountingReader struct {
	r io.Reader
	n atomic.Int64
}

// NewCountingReader returns a CountingReader wrapping r.
func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{r: r}
}

// Read implements io.Reader.
func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n.Add(int64(n))
	return n, err
}

// N returns the total number of bytes read so far.
func (c *CountingReader) N() int64 {
	return c.n.Load()
}

// CountingWriter wraps an io.Writer and counts the total number of bytes
// successfully written. It is safe for concurrent use.
type CountingWriter struct {
	w io.Writer
	n atomic.Int64
}

// NewCountingWriter returns a CountingWriter wrapping w.
func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{w: w}
}

// Write implements io.Writer.
func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n.Add(int64(n))
	return n, err
}

// N returns the total number of bytes written so far.
func (c *CountingWriter) N() int64 {
	return c.n.Load()
}
