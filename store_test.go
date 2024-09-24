package main

import (
	"bytes"
	"fmt"
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
	s := newStore()
	defer teardown(t, s)

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")

		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, _ := io.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf("expected to NOT have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransform,
		Root:              "distributed_storage",
	}
	return NewStore(&opts)
}
func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
