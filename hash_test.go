package io2

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"testing"
)

func TestHashReader(t *testing.T) {
	const input = "the quick brown fox"
	hr := NewHashReader(strings.NewReader(input), sha256.New())

	got, err := io.ReadAll(hr)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != input {
		t.Errorf("bytes = %q; want %q", got, input)
	}

	want := sha256.Sum256([]byte(input))
	if gotHex, wantHex := hex.EncodeToString(hr.Sum(nil)), hex.EncodeToString(want[:]); gotHex != wantHex {
		t.Errorf("sum = %s; want %s", gotHex, wantHex)
	}
}
