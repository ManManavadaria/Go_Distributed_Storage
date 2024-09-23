package main

import (
	"bytes"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "HelloWorld"

	pathnkey := CASPathTransform(key)

	expectedPathname := "db8ac/1c259/eb89d/4a131/b253b/acfca/5f319/d54f2"
	expectedOriginalKey := "db8ac1c259eb89d4a131b253bacfca5f319d54f2"

	if pathnkey.PathName != expectedPathname {
		t.Errorf("Got %s want %s", pathnkey.PathName, expectedPathname)
	}
	if pathnkey.FileName != expectedOriginalKey {
		t.Errorf("Got %s want %s", pathnkey.FileName, expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {
	opts := &StoreOpts{
		PathTransformFunc: CASPathTransform,
	}

	key := "HelloWorld"

	s := NewStore(opts)

	data := []byte("Test Bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if exist := s.Has(key); !exist {
		t.Errorf("Want %v have %v", true, exist)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err.Error())
	}

	b, _ := io.ReadAll(r)

	if string(b) != string(data) {
		t.Errorf("Want %s have %s", data, b)
	}
}
func TestDelete(t *testing.T) {
	opts := &StoreOpts{
		PathTransformFunc: CASPathTransform,
	}

	key := "HelloWorld"

	s := NewStore(opts)

	data := []byte("Test Bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestHas(t *testing.T) {
	opts := &StoreOpts{
		PathTransformFunc: CASPathTransform,
	}

	key := "Has Test Key"

	s := NewStore(opts)

	if exist := s.Has(key); exist {
		t.Errorf("Want %v have %v", false, exist)
	}
}
