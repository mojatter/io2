# github.com/mojatter/io2

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mojatter/io2)](https://pkg.go.dev/github.com/mojatter/io2)
[![Report Card](https://goreportcard.com/badge/github.com/mojatter/io2)](https://goreportcard.com/report/github.com/mojatter/io2)
[![Coverage Status](https://coveralls.io/repos/github/mojatter/io2/badge.svg?branch=main)](https://coveralls.io/github/mojatter/io2?branch=main)

Go "io" package utilities.

- [Delegator](#delegator)
- [No-op Closer](#no-op-closer)
- [WriteSeeker](#writeseeker)
- [Multi Readers](#multi-readers)

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

```go
// NopReadCloser returns a ReadCloser with a no-op Close method wrapping the provided interface.
// This function like io.NopCloser(io.Reader).
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
