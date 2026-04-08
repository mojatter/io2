package io2

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// MultiReader represents a multiple reader.
type MultiReader interface {
	io.Reader
	// Current returns a current index of multiple readers. The current starts 0.
	Current() int
}

// MultiReadCloser is the interface that groups the MultiReader and Close methods.
type MultiReadCloser interface {
	MultiReader
	io.Closer
}

// MultiReadSeeker is the interface that groups the MultiReader, Seek and SeekReader methods.
type MultiReadSeeker interface {
	MultiReader
	io.Seeker
	// SeekReader sets the offset of multiple readers.
	SeekReader(current int) (int64, error)
}

// MultiReadSeekCloser is the interface that groups the MultiReadSeeker and Close methods.
type MultiReadSeekCloser interface {
	MultiReadSeeker
	io.Closer
}

var errSkip = errors.New("skip")

type singleReader struct {
	io.ReadSeekCloser
	off    int64
	length int64
}

type multiReader struct {
	rs      []*singleReader
	current int
	length  int64
}

var _ io.ReadSeekCloser = (*multiReader)(nil)

// NewMultiReader creates a Reader that's the logical concatenation
// of the provided input readers.
func NewMultiReader(rs ...io.Reader) MultiReader {
	ds := make([]*singleReader, len(rs))
	for i, r := range rs {
		ds[i] = &singleReader{ReadSeekCloser: Delegate(r)}
	}
	return &multiReader{rs: ds}
}

// NewMultiReadCloser create a ReaderCloser that's the logical concatenation
// of the provided input readers.
func NewMultiReadCloser(rs ...io.ReadCloser) MultiReadCloser {
	ds := make([]*singleReader, len(rs))
	for i, r := range rs {
		ds[i] = &singleReader{ReadSeekCloser: Delegate(r)}
	}
	return &multiReader{rs: ds}
}

// NewMultiReadSeeker creates a ReadSeeker that's the logical concatenation
// of the provided input readers.
func NewMultiReadSeeker(rs ...io.ReadSeeker) (MultiReadSeeker, error) {
	ds := make([]io.ReadSeekCloser, len(rs))
	for i, r := range rs {
		ds[i] = NopReadSeekCloser(r)
	}
	return NewMultiReadSeekCloser(ds...)
}

// NewMultiReadSeekCloser creates a ReadSeekCloser that's the logical
// concatenation of the provided input readers.
func NewMultiReadSeekCloser(rs ...io.ReadSeekCloser) (MultiReadSeekCloser, error) {
	length := int64(0)
	ds := make([]*singleReader, len(rs))
	for i, r := range rs {
		n, err := r.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		if _, err = r.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		ds[i] = &singleReader{
			ReadSeekCloser: Delegate(r),
			length:         n,
		}
		length += n
	}
	return &multiReader{rs: ds, length: length}, nil
}

func NewMultiStringReader(strs ...string) MultiReadSeeker {
	length := int64(0)
	ds := make([]*singleReader, len(strs))
	for i, str := range strs {
		ds[i] = &singleReader{
			ReadSeekCloser: NopReadSeekCloser(strings.NewReader(str)),
			length:         int64(len(str)),
		}
		length += ds[i].length
	}
	return &multiReader{rs: ds, length: length}
}

func NewMultiFileReader(filenames ...string) (MultiReadSeekCloser, error) {
	length := int64(0)
	ds := make([]*singleReader, len(filenames))
	for i, filename := range filenames {
		f, err := osOpen(filename)
		if err != nil {
			return nil, err
		}
		info, err := fsStat(f)
		if err != nil {
			return nil, err
		}
		ds[i] = &singleReader{
			ReadSeekCloser: f,
			length:         info.Size(),
		}
		length += ds[i].length
	}
	return &multiReader{rs: ds, length: length}, nil
}

// Current returns a current index of multiple readers.
func (mr *multiReader) Current() int {
	return mr.current
}

func (mr *multiReader) each(i int, offset int64, fn func(r *singleReader) error) error {
	mr.current = i
	if offset >= 0 {
		for ; mr.current < len(mr.rs); mr.current++ {
			if err := fn(mr.rs[mr.current]); err != nil {
				if err == errSkip {
					err = nil
				}
				return err
			}
		}
		if mr.current >= len(mr.rs) {
			mr.current = len(mr.rs) - 1
		}
		return nil
	}
	for ; mr.current >= 0; mr.current-- {
		if err := fn(mr.rs[mr.current]); err != nil {
			if err == errSkip {
				err = nil
			}
			return err
		}
	}
	if mr.current < 0 {
		mr.current = 0
	}
	return nil
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
	if mr.current >= len(mr.rs) {
		return 0, io.EOF
	}
	off := 0
	for ; mr.current < len(mr.rs); mr.current++ {
		r := mr.rs[mr.current]
		n, err = r.Read(p[off:])
		if err != nil {
			return 0, err
		}
		r.off += int64(n)
		off += n
		if off >= len(p) {
			return off, nil
		}
	}
	return off, nil
}

func (mr *multiReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		return mr.seekStart(offset, 0)
	case io.SeekCurrent:
		return mr.seekCurrent(offset)
	case io.SeekEnd:
		return mr.seekEnd(offset, len(mr.rs)-1)
	}
	return 0, errors.New("invalid whence")
}

// SeekReader sets the offset of multiple readers. The current starts 0.
func (mr *multiReader) SeekReader(current int) (int64, error) {
	var offset int64
	for i, r := range mr.rs {
		if i >= current {
			break
		}
		offset += r.length
	}
	return mr.Seek(offset, io.SeekStart)
}

func (mr *multiReader) resetTails() error {
	for i := mr.current + 1; i < len(mr.rs); i++ {
		r := mr.rs[i]
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return err
		}
		r.off = 0
	}
	return nil
}

func (mr *multiReader) seekStart(offset int64, start int) (int64, error) {
	off := int64(0)
	for i := 0; i < start; i++ {
		off += mr.rs[i].length
	}
	err := mr.each(start, offset, func(r *singleReader) error {
		safeOffset := offset
		if mr.current != len(mr.rs)-1 && safeOffset > (r.length-r.off) {
			safeOffset = (r.length - r.off)
		}
		n, err := r.Seek(safeOffset, io.SeekStart)
		if err != nil {
			return err
		}
		r.off = n
		if offset < r.length {
			off += offset
			return errSkip
		}
		offset -= safeOffset
		off += safeOffset
		return nil
	})
	if err != nil {
		return 0, err
	}
	if err := mr.resetTails(); err != nil {
		return 0, err
	}
	return off, nil
}

func (mr *multiReader) seekCurrent(offset int64) (int64, error) {
	off := int64(0)
	for i := 0; i < mr.current; i++ {
		off += mr.rs[i].length
	}

	r := mr.rs[mr.current]
	safeOffset := offset
	if offset >= 0 {
		if mr.current != len(mr.rs)-1 && safeOffset > (r.length-r.off) {
			safeOffset = (r.length - r.off)
		}
	} else {
		if mr.current != 0 && -safeOffset > r.off {
			safeOffset = -r.off
		}
	}
	n, err := r.Seek(safeOffset, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	r.off = n
	if diffset := offset - safeOffset; diffset != 0 {
		if offset >= 0 {
			return mr.seekStart(diffset, mr.current+1)
		}
		return mr.seekEnd(diffset, mr.current-1)
	}
	if err := mr.resetTails(); err != nil {
		return 0, err
	}
	return off + n, nil
}

func (mr *multiReader) seekEnd(offset int64, end int) (int64, error) {
	off := mr.length
	for i := end + 1; i < len(mr.rs); i++ {
		off -= mr.rs[i].length
	}
	err := mr.each(end, offset, func(r *singleReader) error {
		safeOffset := offset
		if mr.current != 0 && -safeOffset > r.length {
			safeOffset = -r.length
		}
		n, err := r.Seek(safeOffset, io.SeekEnd)
		if err != nil {
			return err
		}
		r.off = n
		if -offset < r.length {
			off += offset
			return errSkip
		}
		offset += r.length
		off -= r.length
		return nil
	})
	if err != nil {
		return 0, err
	}
	return off, nil
}

func (mr *multiReader) Close() error {
	var errs []string
	// Callback never returns an error; Close failures are aggregated in errs.
	_ = mr.each(0, 1, func(r *singleReader) error {
		if err := r.Close(); err != nil {
			errs = append(errs, err.Error())
		}
		return nil
	})
	if len(errs) > 0 {
		return fmt.Errorf("failed to close: %s", strings.Join(errs, "; "))
	}
	mr.rs = nil
	mr.current = 0
	mr.length = 0
	return nil
}
