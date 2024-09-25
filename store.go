package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(string) PathKey

var CASPathTransform PathTransformFunc = func(key string) PathKey {
	hash := sha1.Sum([]byte(key))

	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5

	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

type PathKey struct {
	PathName string
	FileName string
}

type StoreOpts struct {
	Root              string
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

func NewStore(str *StoreOpts) *Store {
	if str.PathTransformFunc == nil {
		str.PathTransformFunc = DefaultPathTransformFunc
	}
	return &Store{
		StoreOpts: *str,
	}
}

func (p *PathKey) FullPath(root string) string {
	return fmt.Sprintf("%s/%s/%s", root, p.PathName, p.FileName)
}

func (s *Store) Has(key string) bool {
	pathkey := s.PathTransformFunc(key)

	// fmt.Println("\033[32m", pathkey, "\033[0m")

	_, err := os.Stat(pathkey.FullPath(s.Root))
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}

	if err != nil {
		fmt.Println("Error occurred:", err)
		return false
	}

	return true
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathTransformFunc(key)

	defer func() {
		fmt.Printf("Deleted [%s] from disk", pathkey.FullPath(s.Root))
	}()

	path := s.Root + "/" + strings.Split(pathkey.PathName, "/")[0]

	return os.RemoveAll(path)
}

func (s *Store) Read(key string) (io.Reader, error) {

	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathkey := s.PathTransformFunc(key)

	f, err := os.Open(pathkey.FullPath(s.Root))
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathkey := s.PathTransformFunc(key)

	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return err
	}

	fullpath := pathkey.FullPath(s.Root)

	f, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("Written (%d) bytes to disk: %s", n, fullpath)

	return nil
}
