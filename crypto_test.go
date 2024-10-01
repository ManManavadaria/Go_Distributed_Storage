package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo Bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := NewEncryptionKey()

	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(len(payload))
	fmt.Println(len(key))
	fmt.Println(len(dst.String()))

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(nw)
	fmt.Println(out.String())

	if nw != 16+len(payload) {
		t.Fail()
	}

	if out.String() != payload {
		t.Error("decryption failed")
	}
}
