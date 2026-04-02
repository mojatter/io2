package io2

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testMultiFilenames(contents ...string) ([]string, func(), error) {
	dir, err := os.MkdirTemp("", "*.multireader")
	if err != nil {
		return nil, nil, err
	}
	done := func() { os.RemoveAll(dir) }
	var names []string
	for i, data := range contents {
		name := filepath.Join(dir, fmt.Sprintf("%d.txt", i+1))
		if err = os.WriteFile(name, []byte(data), os.ModePerm); err != nil {
			break
		}
		names = append(names, name)
	}
	if err != nil {
		done()
		return nil, nil, err
	}
	return names, done, nil
}

func mustNewMultiFileReader(t *testing.T, filenames ...string) MultiReadSeekCloser {
	r, err := NewMultiFileReader(filenames...)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestMultiRead(t *testing.T) {
	tests := []struct {
		reader io.ReadCloser
		n      int
		errstr string
		want   string
	}{
		{
			reader: NopReadCloser(NewMultiReader(
				strings.NewReader("abc"),
				strings.NewReader("def"),
			)),
			n:    6,
			want: "abcdef",
		}, {
			reader: NopReadCloser(NewMultiStringReader("abc", "def")),
			n:      6,
			want:   "abcdef",
		}, {
			reader: NewMultiReadCloser(&Delegator{
				ReadFunc: func(p []byte) (int, error) {
					return 0, errors.New("failed to read for coverage")
				},
			}),
			errstr: "failed to read for coverage",
		},
	}

	var deferClose func() error
	close := func() {
		if deferClose != nil {
			deferClose()
		}
	}
	defer close()

	for i, test := range tests {
		close()
		r := test.reader
		deferClose = r.Close

		p := make([]byte, test.n)
		n, err := r.Read(p)
		if test.errstr != "" {
			if err == nil {
				t.Fatalf("tests[%d] no error", i)
			}
			if err.Error() != test.errstr {
				t.Errorf("tests[%d] error %s; want %s", i, err.Error(), test.errstr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("tests[%d] error %v", i, err)
		}
		if n != test.n {
			t.Errorf("tests[%d] n is %d; want %d", i, n, test.n)
		}
		got := string(p)
		if got != test.want {
			t.Errorf("tests[%d] got %s; want %s", i, got, test.want)
		}
	}
}

func TestMultiSeek(t *testing.T) {
	filenames, done, err := testMultiFilenames("abc", "def", "ghi")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	newErrSeekReaders := func() []io.ReadSeekCloser {
		return []io.ReadSeekCloser{&multiReader{
			rs: []*singleReader{
				{
					ReadSeekCloser: &Delegator{
						SeekFunc: func(offset int64, whence int) (int64, error) {
							return 0, errors.New("failed to seek for coverage")
						},
					},
				},
			},
		}}
	}

	tests := []struct {
		readers func() []io.ReadSeekCloser
		offset  int64
		whence  int
		n       int64
		errstr  string
		after   string
	}{
		{
			readers: func() []io.ReadSeekCloser {
				return []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
			},
			offset: 2,
			whence: io.SeekStart,
			n:      2,
			after:  "cdefghi",
		}, {
			readers: func() []io.ReadSeekCloser {
				return []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
			},
			offset: 4,
			whence: io.SeekStart,
			n:      4,
			after:  "efghi",
		}, {
			readers: func() []io.ReadSeekCloser {
				rs := []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
				for _, r := range rs {
					io.ReadAll(r)
				}
				return rs
			},
			offset: 0,
			whence: io.SeekStart,
			n:      0,
			after:  "abcdefghi",
		}, {
			readers: func() []io.ReadSeekCloser {
				r0 := NopReadSeekCloser(strings.NewReader("abc"))
				r1 := strings.NewReader("def")
				d1 := Delegate(r1)
				r, err := NewMultiReadSeekCloser(r0, d1)
				if err != nil {
					t.Fatal(err)
				}
				d1.SeekFunc = func(offset int64, whence int) (int64, error) {
					return 0, errors.New("failed to reset tails")
				}
				return []io.ReadSeekCloser{r}
			},
			offset: 0,
			whence: io.SeekStart,
			errstr: "failed to reset tails",
		}, {
			readers: func() []io.ReadSeekCloser {
				return []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
			},
			offset: -4,
			whence: io.SeekEnd,
			n:      5,
			after:  "fghi",
		}, {
			readers: func() []io.ReadSeekCloser {
				return []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
			},
			offset: -9,
			whence: io.SeekEnd,
			n:      0,
			after:  "abcdefghi",
		}, {
			readers: func() []io.ReadSeekCloser {
				rs := []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
				for _, r := range rs {
					if _, err := r.Seek(4, io.SeekStart); err != nil {
						t.Fatal(err)
					}
				}
				return rs
			},
			offset: 3,
			whence: io.SeekCurrent,
			n:      7,
			after:  "hi",
		}, {
			readers: func() []io.ReadSeekCloser {
				rs := []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
				for _, r := range rs {
					if _, err := r.Seek(4, io.SeekStart); err != nil {
						t.Fatal(err)
					}
				}
				return rs
			},
			offset: -1,
			whence: io.SeekCurrent,
			n:      3,
			after:  "defghi",
		}, {
			readers: func() []io.ReadSeekCloser {
				rs := []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
				for _, r := range rs {
					if _, err := r.Seek(4, io.SeekStart); err != nil {
						t.Fatal(err)
					}
				}
				return rs
			},
			offset: -4,
			whence: io.SeekCurrent,
			n:      0,
			after:  "abcdefghi",
		}, {
			readers: func() []io.ReadSeekCloser {
				r0 := NopReadSeekCloser(strings.NewReader("abc"))
				r1 := strings.NewReader("def")
				d1 := Delegate(r1)
				r, err := NewMultiReadSeekCloser(r0, d1)
				if err != nil {
					t.Fatal(err)
				}
				d1.SeekFunc = func(offset int64, whence int) (int64, error) {
					return 0, errors.New("failed to reset tails")
				}
				return []io.ReadSeekCloser{r}
			},
			offset: 0,
			whence: io.SeekCurrent,
			errstr: "failed to reset tails",
		}, {
			readers: func() []io.ReadSeekCloser {
				return []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
				}
			},
			offset: -1,
			whence: io.SeekStart,
			errstr: "strings.Reader.Seek: negative position",
		}, {
			readers: func() []io.ReadSeekCloser {
				return []io.ReadSeekCloser{
					NopReadSeekCloser(strings.NewReader("abcdefghi")),
					NopReadSeekCloser(NewMultiStringReader("abc", "def", "ghi")),
					mustNewMultiFileReader(t, filenames...),
				}
			},
			offset: 10,
			whence: io.SeekStart,
			n:      10,
		}, {
			readers: func() []io.ReadSeekCloser {
				return []io.ReadSeekCloser{NopReadSeekCloser(NewMultiStringReader("abc", "def"))}
			},
			whence: -1,
			errstr: "invalid whence",
		}, {
			readers: newErrSeekReaders,
			whence:  io.SeekStart,
			errstr:  "failed to seek for coverage",
		}, {
			readers: newErrSeekReaders,
			whence:  io.SeekCurrent,
			errstr:  "failed to seek for coverage",
		}, {
			readers: newErrSeekReaders,
			whence:  io.SeekEnd,
			errstr:  "failed to seek for coverage",
		},
	}

	fn0 := func(i, j int, r io.ReadSeekCloser) {
		defer r.Close()

		test := tests[i]
		n, err := r.Seek(test.offset, test.whence)
		if test.errstr != "" {
			if err == nil {
				t.Fatalf("tests[%d-%d] no error", i, j)
			}
			if err.Error() != test.errstr {
				t.Errorf("tests[%d-%d] error %s; want %s", i, j, err.Error(), test.errstr)
			}
			return
		}
		if err != nil {
			t.Fatalf("tests[%d-%d] error %v", i, j, err)
		}
		if n != test.n {
			t.Errorf("tests[%d-%d] n is %d; want %d", i, j, n, test.n)
		}

		p, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("tests[%d-%d] error %v", i, j, err)
		}
		after := string(p)
		if after != test.after {
			t.Errorf("tests[%d-%d] after %s; want %s", i, j, after, test.after)
		}
	}
	fn1 := func(i int) {
		test := tests[i]
		rs := test.readers()
		for j, r := range rs {
			fn0(i, j, r)
		}
	}
	for i := range tests {
		fn1(i)
	}
}

func TestMultiReadSeeker_Errors(t *testing.T) {
	tests := []struct {
		r      io.ReadSeeker
		errstr string
	}{
		{
			r: &Delegator{
				SeekFunc: func(offset int64, whence int) (int64, error) {
					if whence == io.SeekStart {
						return 0, errors.New("failed to seek start for coverage")
					}
					return 0, nil
				},
			},
			errstr: "failed to seek start for coverage",
		}, {
			r: &Delegator{
				SeekFunc: func(offset int64, whence int) (int64, error) {
					if whence == io.SeekEnd {
						return 0, errors.New("failed to seek end for coverage")
					}
					return 0, nil
				},
			},
			errstr: "failed to seek end for coverage",
		},
	}
	for i, test := range tests {
		_, err := NewMultiReadSeeker(test.r)
		if err == nil {
			t.Fatalf("tests[%d] no error", i)
		}
		if err.Error() != test.errstr {
			t.Errorf("tests[%d] error %s; want %s", i, err.Error(), test.errstr)
		}
	}
}

func TestMultiClose(t *testing.T) {
	errCloseReader := &Delegator{
		CloseFunc: func() error {
			return errors.New("close error")
		},
	}

	tests := []struct {
		r      io.Closer
		errstr string
	}{
		{
			r: NewMultiReadCloser(NopReadCloser(strings.NewReader("no error"))),
		}, {
			r:      NewMultiReadCloser(errCloseReader),
			errstr: "failed to close: close error",
		}, {
			r:      NewMultiReadCloser(errCloseReader, errCloseReader),
			errstr: "failed to close: close error; close error",
		},
	}
	for i, test := range tests {
		err := test.r.Close()
		if test.errstr != "" {
			if err == nil {
				t.Fatalf("tests[%d] no error", i)
			}
			if err.Error() != test.errstr {
				t.Errorf("tests[%d] error %s; want %s", i, err.Error(), test.errstr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("tests[%d] error %v", i, err)
		}
	}
}

func TestMultiSeekReader(t *testing.T) {
	tests := []struct {
		reader  func() MultiReadSeeker
		offset  int
		n       int64
		current int
	}{
		{
			reader: func() MultiReadSeeker {
				return NewMultiStringReader("a", "bc", "def")
			},
			offset:  0,
			n:       0,
			current: 0,
		}, {
			reader: func() MultiReadSeeker {
				return NewMultiStringReader("a", "bc", "def")
			},
			offset:  2,
			n:       3,
			current: 2,
		}, {
			reader: func() MultiReadSeeker {
				return NewMultiStringReader("a", "bc", "def")
			},
			offset:  10,
			n:       6,
			current: 2,
		}, {
			reader: func() MultiReadSeeker {
				r := NewMultiStringReader("a", "bc", "def")
				r.Seek(2, io.SeekStart)
				return r
			},
			offset:  0,
			n:       0,
			current: 0,
		}, {
			reader: func() MultiReadSeeker {
				r := NewMultiStringReader("a", "bc", "def")
				r.Seek(5, io.SeekStart)
				return r
			},
			offset:  1,
			n:       1,
			current: 1,
		},
	}
	for i, test := range tests {
		r := test.reader()
		n, err := r.SeekReader(test.offset)
		if err != nil {
			t.Fatalf("tests[%d] error %v", i, err)
		}
		if n != test.n {
			t.Errorf("tests[%d] n is %d; want %d", i, n, test.n)
		}
		if current := r.Current(); current != test.current {
			t.Errorf("tests[%d] current is %d; want %d", i, current, test.current)
		}
	}
}

func TestNewMultiFileReader_osOpenError(t *testing.T) {
	osOpenOrg := osOpen
	defer func() { osOpen = osOpenOrg }()
	osOpen = func(filename string) (*os.File, error) {
		return nil, errors.New("test-error")
	}

	_, err := NewMultiFileReader("LICENSE")
	if err == nil || err.Error() != "test-error" {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestNewMultiFileReader_fsStatError(t *testing.T) {
	fsStatOrg := fsStat
	defer func() { fsStat = fsStatOrg }()
	fsStat = func(file *os.File) (fs.FileInfo, error) {
		return nil, errors.New("test-error")
	}

	_, err := NewMultiFileReader("LICENSE")
	if err == nil || err.Error() != "test-error" {
		t.Fatalf("unexpected error %v", err)
	}
}
