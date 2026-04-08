package io2

import (
	"io"
)

// Delegator implements Reader, Writer, Seeker, Closer.
type Delegator struct {
	ReadFunc  func(p []byte) (n int, err error)
	WriteFunc func(p []byte) (n int, err error)
	SeekFunc  func(offset int64, whence int) (int64, error)
	CloseFunc func() error
}

var (
	_ io.Reader = (*Delegator)(nil)
	_ io.Writer = (*Delegator)(nil)
	_ io.Seeker = (*Delegator)(nil)
	_ io.Closer = (*Delegator)(nil)
)

// Read calls ReadFunc(p).
func (d *Delegator) Read(p []byte) (int, error) {
	if d.ReadFunc == nil {
		return 0, ErrNotImplemented
	}
	return d.ReadFunc(p)
}

// Write calls WriteFunc(p).
func (d *Delegator) Write(p []byte) (int, error) {
	if d.WriteFunc == nil {
		return 0, ErrNotImplemented
	}
	return d.WriteFunc(p)
}

// Seek calls SeekFunc(offset, whence).
func (d *Delegator) Seek(offset int64, whence int) (int64, error) {
	if d.SeekFunc == nil {
		return 0, ErrNotImplemented
	}
	return d.SeekFunc(offset, whence)
}

// Close calls CloseFunc().
func (d *Delegator) Close() error {
	if d.CloseFunc == nil {
		// NOTE: return no error.
		return nil
	}
	return d.CloseFunc()
}

// Delegate returns a Delegator with the provided io interfaces (io.Reader, io.Seeker, io.Writer, io.Closer).
func Delegate(i interface{}) *Delegator {
	d := &Delegator{}
	if r, ok := i.(io.Reader); ok {
		d.ReadFunc = r.Read
	}
	if s, ok := i.(io.Seeker); ok {
		d.SeekFunc = s.Seek
	}
	if w, ok := i.(io.Writer); ok {
		d.WriteFunc = w.Write
	}
	if c, ok := i.(io.Closer); ok {
		d.CloseFunc = c.Close
	}
	return d
}

// DelegateReader returns a Delegator with the provided Read function.
func DelegateReader(i io.Reader) *Delegator {
	return &Delegator{
		ReadFunc: i.Read,
	}
}

// DelegateReadCloser returns a Delegator with the provided Read and Close functions.
func DelegateReadCloser(i io.ReadCloser) *Delegator {
	return &Delegator{
		ReadFunc:  i.Read,
		CloseFunc: i.Close,
	}
}

// DelegateReadSeeker returns a Delegator with the provided Read and Seek functions.
func DelegateReadSeeker(i io.ReadSeeker) *Delegator {
	return &Delegator{
		ReadFunc: i.Read,
		SeekFunc: i.Seek,
	}
}

// DelegateReadSeekCloser returns a Delegator with the provided Read, Seek and Close functions.
func DelegateReadSeekCloser(i io.ReadSeekCloser) *Delegator {
	return &Delegator{
		ReadFunc:  i.Read,
		SeekFunc:  i.Seek,
		CloseFunc: i.Close,
	}
}

// DelegateReadWriteCloser returns a Delegator with the provided Read, Write and Close functions.
func DelegateReadWriteCloser(i io.ReadWriteCloser) *Delegator {
	return &Delegator{
		ReadFunc:  i.Read,
		WriteFunc: i.Write,
		CloseFunc: i.Close,
	}
}

// DelegateReadWriteSeeker returns a Delegator with the provided Read, Write and Seek functions.
func DelegateReadWriteSeeker(i io.ReadWriteSeeker) *Delegator {
	return &Delegator{
		ReadFunc:  i.Read,
		WriteFunc: i.Write,
		SeekFunc:  i.Seek,
	}
}

// DelegateReadWriter returns a Delegator with the provided Read and Write functions.
func DelegateReadWriter(i io.ReadWriter) *Delegator {
	return &Delegator{
		ReadFunc:  i.Read,
		WriteFunc: i.Write,
	}
}

// DelegateWriter returns a Delegator with the provided Write function.
func DelegateWriter(i io.Writer) *Delegator {
	return &Delegator{
		WriteFunc: i.Write,
	}
}

// DelegateWriteCloser returns a Delegator with the provided Write and Close functions.
func DelegateWriteCloser(i io.WriteCloser) *Delegator {
	return &Delegator{
		WriteFunc: i.Write,
		CloseFunc: i.Close,
	}
}

// DelegateWriteSeeker returns a Delegator with the provided Write and Seek functions.
func DelegateWriteSeeker(i io.WriteSeeker) *Delegator {
	return &Delegator{
		WriteFunc: i.Write,
		SeekFunc:  i.Seek,
	}
}

// DelegateWriteSeekCloser returns a Delegator with the provided Write, Seek and Close functions.
func DelegateWriteSeekCloser(i WriteSeekCloser) *Delegator {
	return &Delegator{
		WriteFunc: i.Write,
		SeekFunc:  i.Seek,
		CloseFunc: i.Close,
	}
}

// NopReadCloser returns a ReadCloser with a no-op Close method wrapping the provided interface.
//
// Deprecated: use the standard library's io.NopCloser instead.
func NopReadCloser(r io.Reader) io.ReadCloser {
	return DelegateReader(r)
}

// NopReadWriteCloser returns a ReadWriteCloser with a no-op Close method wrapping the provided interface.
func NopReadWriteCloser(rw io.ReadWriter) io.ReadWriteCloser {
	return DelegateReadWriter(rw)
}

// NopReadSeekCloser returns a ReadSeekCloser with a no-op Close method wrapping the provided interface.
func NopReadSeekCloser(r io.ReadSeeker) io.ReadSeekCloser {
	return DelegateReadSeeker(r)
}

// NopWriteCloser returns a WriteCloser with a no-op Close method wrapping the provided interface.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return DelegateWriter(w)
}
