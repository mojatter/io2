# github.com/mojatter/io2

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mojatter/io2)](https://pkg.go.dev/github.com/mojatter/io2)
[![Report Card](https://goreportcard.com/badge/github.com/mojatter/io2)](https://goreportcard.com/report/github.com/mojatter/io2)
[![Coverage Status](https://coveralls.io/repos/github/mojatter/io2/badge.svg?branch=main)](https://coveralls.io/github/mojatter/io2?branch=main)

Go "io" package utilities.

- [Delegator](#delegator)
- [No-op Closer](#no-op-closer)
- [WriteSeeker](#writeseeker)
- [Multi Readers](#multi-readers)
- [Counting Reader / Writer](#counting-reader--writer)
- [Hash Reader](#hash-reader)
- [Context Reader / Writer](#context-reader--writer)

## Delegator

Delegator implements io.Reader, io.Writer, io.Seeker, io.Closer.
Delegator can override the I/O functions that is useful for unit tests.

```go
package main

import (
  "bytes"
  "errors"
  "fmt"
  "io/ioutil"

  "github.com/mojatter/io2"
)

func main() {
  org := bytes.NewReader([]byte(`original`))

  r := io2.DelegateReader(org)
  r.ReadFunc = func(p []byte) (int, error) {
    return 0, errors.New("custom")
  }

  var err error
  _, err = ioutil.ReadAll(r)
  fmt.Printf("Error: %v\n", err)

  // Output: Error: custom
}
```

### No-op Closer

> **Note:** `NopReadCloser` is deprecated since v0.9.0; use the standard library's
> `io.NopCloser` instead. The other helpers remain useful because `io.NopCloser`
> only accepts a plain `io.Reader`.

```go
// NopReadCloser returns a ReadCloser with a no-op Close method wrapping the provided interface.
// Deprecated: use io.NopCloser.
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
```

## WriteSeeker

WriteSeekBuffer implements io.Writer, io.Seeker and io.Closer.
NewWriteSeekBuffer(capacity int) returns the buffer.

```go
// WriteSeekCloser is the interface that groups the basic Write, Seek and Close methods.
type WriteSeekCloser interface {
  io.Writer
  io.Seeker
  io.Closer
}
```

```go
package main

import (
  "fmt"
  "io"

  "github.com/mojatter/io2"
)

func main() {
  o := io2.NewWriteSeekBuffer(16)
  o.Write([]byte(`Hello!`))
  o.Truncate(o.Len() - 1)
  o.Write([]byte(` world!`))

  fmt.Println(string(o.Bytes()))

  o.Seek(-1, io.SeekEnd)
  o.Write([]byte(`?`))

  fmt.Println(string(o.Bytes()))

  // Output:
  // Hello world!
  // Hello world?
}
```

## Multi Readers

io2 provides MultiReadCloser, MultiReadSeeker, MultiReadSeekCloser.

```go
package main

import (
  "fmt"
  "io"
  "io/ioutil"
  "strings"

  "github.com/mojatter/io2"
)

func main() {
  r, _ := io2.NewMultiReadSeeker(
    strings.NewReader("Hello !"),
    strings.NewReader(" World"),
  )

  r.Seek(5, io.SeekStart)
  p, _ := ioutil.ReadAll(r)
  fmt.Println(string(p))

  r.Seek(-5, io.SeekEnd)
  p, _ = ioutil.ReadAll(r)
  fmt.Println(string(p))

  // Output:
  // ! World
  // World
}
```

## Counting Reader / Writer

`CountingReader` and `CountingWriter` track the total number of bytes
transferred through them. They are safe for concurrent use and useful for
progress reporting.

```go
cr := io2.NewCountingReader(resp.Body)
io.Copy(dst, cr)
fmt.Printf("read %d bytes\n", cr.N())
```

## Hash Reader

`HashReader` updates a `hash.Hash` with every byte read, so you can verify or
fingerprint content while streaming it elsewhere — for example computing an S3
ETag while uploading.

```go
hr := io2.NewHashReader(file, sha256.New())
io.Copy(uploader, hr)
fmt.Printf("sha256=%x\n", hr.Sum(nil))
```

## Context Reader / Writer

`NewContextReader` and `NewContextWriter` abort reads and writes once the
supplied `context.Context` is canceled. The context is checked before each
call, so they work best with readers and writers that return control
frequently (network, buffered I/O).

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

r := io2.NewContextReader(ctx, conn)
_, err := io.Copy(dst, r) // returns context.DeadlineExceeded on timeout
```
