package io2

import (
	"io"
	"reflect"
	"testing"
)

func TestWrite(t *testing.T) {
	tests := []struct {
		capacity  int
		p         []byte
		wantBytes []byte
		wantOff   int
		wantLen   int
	}{
		{
			capacity:  8,
			p:         []byte(`123`),
			wantBytes: []byte{'1', '2', '3'},
			wantOff:   3,
			wantLen:   3,
		}, {
			p:         []byte(`456`),
			wantBytes: []byte{'1', '2', '3', '4', '5', '6'},
			wantOff:   6,
			wantLen:   6,
		}, {
			p:         []byte(`789`),
			wantBytes: []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9'},
			wantOff:   9,
			wantLen:   9,
		},
	}

	var b *WriteSeekBuffer
	close := func() {
		if b != nil {
			b.Close()
		}
	}
	defer close()

	for i, test := range tests {
		if test.capacity > 0 {
			close()
			b = NewWriteSeekBuffer(test.capacity)
		}
		n, err := b.Write(test.p)
		if err != nil {
			t.Fatalf("tests[%d] write: %v", i, err)
		}
		if n != len(test.p) {
			t.Errorf("tests[%d] write bytes %d; want %d", i, n, len(test.p))
		}
		if b.Offset() != test.wantOff {
			t.Errorf("tests[%d] off %d; want %d", i, b.Offset(), test.wantOff)
		}
		if b.Len() != test.wantLen {
			t.Errorf("tests[%d] len %d; want %d", i, b.Len(), test.wantLen)
		}
		if !reflect.DeepEqual(b.Bytes(), test.wantBytes) {
			t.Errorf("tests[%d] bytes %v; want %v", i, b.Bytes(), test.wantBytes)
		}
	}
}

func TestNewWriteSeekBufferEmpty(t *testing.T) {
	b := NewWriteSeekBuffer(16)
	defer b.Close()

	if got := b.Len(); got != 0 {
		t.Errorf("Len() = %d; want 0", got)
	}
	if got := len(b.Bytes()); got != 0 {
		t.Errorf("len(Bytes()) = %d; want 0", got)
	}
	end, err := b.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatalf("Seek: %v", err)
	}
	if end != 0 {
		t.Errorf("Seek(0, SeekEnd) = %d; want 0", end)
	}
}

func TestSeek(t *testing.T) {
	b := NewWriteSeekBufferBytes([]byte(`123456789`))
	defer b.Close()

	tests := []struct {
		off     int64
		whence  int
		wantOff int64
	}{
		{
			off:     0,
			whence:  io.SeekStart,
			wantOff: 0,
		}, {
			off:     int64(-1),
			whence:  io.SeekStart,
			wantOff: 0,
		}, {
			off:     0,
			whence:  io.SeekEnd,
			wantOff: int64(9),
		}, {
			off:     int64(-1),
			whence:  io.SeekCurrent,
			wantOff: int64(8),
		}, {
			off:     int64(-1),
			whence:  io.SeekCurrent,
			wantOff: int64(7),
		}, {
			off:     int64(-3),
			whence:  io.SeekEnd,
			wantOff: int64(6),
		},
	}

	for i, test := range tests {
		n, err := b.Seek(test.off, test.whence)
		if err != nil {
			t.Fatalf("tests[%d] seek: %v", i, err)
		}
		if n != test.wantOff {
			t.Errorf("tests[%d] off %d; want %d", i, n, test.wantOff)
		}
	}
}

func TestSeekWrite(t *testing.T) {
	b := NewWriteSeekBufferBytes([]byte(`123456789`))
	defer b.Close()

	off, err := b.Seek(int64(3), io.SeekStart)
	if err != nil {
		t.Fatalf("seek: %v", err)
	}
	if off != 3 {
		t.Errorf("seek off %d; want %d", off, 3)
	}
	n, err := b.Write([]byte(`def`))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != 3 {
		t.Errorf("write bytes %d; want %d", n, 3)
	}

	want := `123def789`
	got := string(b.Bytes())
	if got != want {
		t.Errorf("bytes %s; want %s", got, want)
	}
}

func TestTruncate(t *testing.T) {
	b := NewWriteSeekBufferBytes([]byte(`123456789`))
	defer b.Close()

	tests := []struct {
		n    int
		want []byte
	}{
		{
			n:    8,
			want: []byte(`12345678`),
		}, {
			n:    -5,
			want: []byte(`123`),
		}, {
			n:    10,
			want: []byte(`123`),
		}, {
			n:    -100,
			want: []byte{},
		},
	}

	for i, test := range tests {
		b.Truncate(test.n)
		got := b.Bytes()
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("tests[%d] truncate bytes %v; want %v", i, got, test.want)
		}
	}
}
