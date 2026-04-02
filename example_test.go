package io2_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mojatter/io2"
)

func ExampleDelegateReader() {
	org := bytes.NewReader([]byte(`original`))

	r := io2.DelegateReader(org)
	r.ReadFunc = func(p []byte) (int, error) {
		return 0, errors.New("custom")
	}

	var err error
	_, err = io.ReadAll(r)
	fmt.Printf("Error: %v\n", err)

	// Output: Error: custom
}

func ExampleNewWriteSeekBuffer() {
	o := io2.NewWriteSeekBuffer(0)
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

func ExampleMultiReadSeeker() {
	r, _ := io2.NewMultiReadSeeker(
		strings.NewReader("Hello !"),
		strings.NewReader(" World"),
	)

	r.Seek(5, io.SeekStart)
	p, _ := io.ReadAll(r)
	fmt.Println(string(p))

	r.Seek(-5, io.SeekEnd)
	p, _ = io.ReadAll(r)
	fmt.Println(string(p))

	// Output:
	// ! World
	// World
}
