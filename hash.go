package io2

import (
	"hash"
	"io"
)

// HashReader wraps an io.Reader and updates a hash.Hash with every byte
// read. After the underlying reader reaches io.EOF, Sum returns the final
// digest.
//
// HashReader is useful for verifying content while streaming it elsewhere,
// for example computing an S3 ETag while uploading.
type HashReader struct {
	r io.Reader
	h hash.Hash
}

// NewHashReader returns a HashReader that reads from r and writes every
// byte read into h.
func NewHashReader(r io.Reader, h hash.Hash) *HashReader {
	return &HashReader{r: r, h: h}
}

// Read implements io.Reader.
func (hr *HashReader) Read(p []byte) (int, error) {
	n, err := hr.r.Read(p)
	if n > 0 {
		// hash.Hash.Write never returns an error.
		_, _ = hr.h.Write(p[:n])
	}
	return n, err
}

// Sum appends the current hash digest to b and returns the resulting slice.
// It is equivalent to calling Sum on the underlying hash.Hash.
func (hr *HashReader) Sum(b []byte) []byte {
	return hr.h.Sum(b)
}

// Hash returns the underlying hash.Hash.
func (hr *HashReader) Hash() hash.Hash {
	return hr.h
}
