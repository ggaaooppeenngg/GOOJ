package main

import (
	"io/ioutil"
	"testing"
)

func TestSimple(t *testing.T) {
	err := SaveFile("test", "test")
	if err != nil {
		t.Fatal(err)
	}
	r, err := GetFile("test")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	out, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "test" {
		t.Fatalf("output not equal get %s\n", out)
	}
}
