package main

import (
	"bytes"
	"os"
	"testing"
)

func TestBoundedExplicitPasswordInput(t *testing.T) {
	previous := os.Stdin
	defer func() { os.Stdin = previous }()
	for _, c := range []struct {
		input  string
		want   []byte
		reject bool
	}{{"  Synthetic password  \n", []byte("  Synthetic password  "), false}, {"Synthetic password\r\n", []byte("Synthetic password"), false}, {"Synthetic password\r", []byte("Synthetic password\r"), false}, {string(bytes.Repeat([]byte("x"), 1025)), nil, true}} {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write([]byte(c.input)); err != nil {
			t.Fatal(err)
		}
		w.Close()
		os.Stdin = r
		got, err := readPassword(true, false)
		r.Close()
		if (err != nil) != c.reject || !c.reject && !bytes.Equal(got, c.want) {
			t.Fatal("stdin delimiter/size policy")
		}
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	os.Stdin = r
	if _, err = readPassword(false, false); err == nil {
		t.Fatal("implicit pipe mode accepted")
	}
}
