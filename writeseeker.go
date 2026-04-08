package io2

import (
	"bytes"
	"io"
)

// WriteSeekCloser is the interface that groups the basic Write, Seek and Close methods.
type WriteSeekCloser interface {
	io.Writer
	io.Seeker
	io.Closer
}

// WriteSeekBuffer implements io.WriteSeeker that using in-memory byte buffer.
type WriteSeekBuffer struct {
	buf *bytes.Buffer
	off int
	len int
}

var _ WriteSeekCloser = (*WriteSeekBuffer)(nil)

// NewWriteSeekBuffer returns an WriteSeekBuffer with the initial capacity.
func NewWriteSeekBuffer(capacity int) *WriteSeekBuffer {
	return &WriteSeekBuffer{
		buf: bytes.NewBuffer(make([]byte, 0, capacity)),
	}
}

// NewWriteSeekBufferBytes returns an WriteSeekBuffer with the initial buffer.
func NewWriteSeekBufferBytes(buf []byte) *WriteSeekBuffer {
	off := len(buf)
	return &WriteSeekBuffer{
		buf: bytes.NewBuffer(buf),
		off: off,
		len: off,
	}
}

// Write appends the contents of p to the buffer, growing the buffer as needed.
// The return value n is the length of p; err is always nil.
func (b *WriteSeekBuffer) Write(p []byte) (int, error) {
	n := len(p)
	noff := b.off + n

	if grow := noff - b.buf.Len(); grow > 0 {
		b.buf.Write(make([]byte, grow))
	}

	copy(b.buf.Bytes()[b.off:noff], p)

	b.off = noff
	if noff > b.len {
		b.len = noff
	}
	return n, nil
}

// Seek sets the offset for the next Write to offset, interpreted according to whence:
//   SeekStart means relative to the start of the file,
//   SeekCurrent means relative to the current offset,
//   SeekEnd means relative to the end.
// Seek returns the new offset relative to the start of the file and an error, if any.
func (b *WriteSeekBuffer) Seek(offset int64, whence int) (int64, error) {
	off := int(offset)
	noff := 0
	switch whence {
	case io.SeekStart:
		noff = off
	case io.SeekCurrent:
		noff = b.off + off
	case io.SeekEnd:
		noff = b.buf.Len() + off
	}
	if noff < 0 {
		noff = 0
	}
	b.off = noff
	return int64(noff), nil
}

// Close calls b.Truncate(0).
func (b *WriteSeekBuffer) Close() error {
	b.Truncate(0)
	return nil
}

// Offset returns the offset.
func (b *WriteSeekBuffer) Offset() int {
	return b.off
}

// Len returns the number of bytes of the buffer; b.Len() == len(b.Bytes()).
func (b *WriteSeekBuffer) Len() int {
	return b.len
}

// Bytes returns a slice of length b.Len() of the buffer.
func (b *WriteSeekBuffer) Bytes() []byte {
	n := b.buf.Len()
	if n > b.len {
		n = b.len
	}
	return b.buf.Bytes()[:n]
}

// Truncate changes the size of the buffer with offset.
func (b *WriteSeekBuffer) Truncate(n int) {
	if n < 0 {
		n = b.off + n
	}
	if n < 0 {
		n = 0
	}
	if n < b.buf.Len() {
		b.buf.Truncate(n)
	}
	b.off = n
	b.len = n
}
