package main

import (
	"bytes"
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
	if pathnkey.Original != expectedOriginalKey {
		t.Errorf("Got %s want %s", pathnkey.Original, expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {
	opts := &StoreOpts{
		PathTransformFunc: CASPathTransform,
	}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("Test Bytes"))

	if err := s.writeStream("TestKey", data); err != nil {
		t.Error(err)
	}
}
