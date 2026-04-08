package io2_test

import (
	"bytes"
	"crypto/sha256"
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

func ExampleCountingReader() {
	cr := io2.NewCountingReader(strings.NewReader("hello world"))
	io.Copy(io.Discard, cr)
	fmt.Printf("read %d bytes\n", cr.N())

	// Output: read 11 bytes
}

func ExampleHashReader() {
	hr := io2.NewHashReader(strings.NewReader("the quick brown fox"), sha256.New())
	io.Copy(io.Discard, hr)
	fmt.Printf("%x\n", hr.Sum(nil))

	// Output: 9ecb36561341d18eb65484e833efea61edc74b84cf5e6ae1b81c63533e25fc8f
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
