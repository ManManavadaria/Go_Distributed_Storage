package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
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
		Original: hashStr,
	}
}

type PathKey struct {
	PathName string
	Original string
}

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStore(str *StoreOpts) *Store {
	return &Store{
		StoreOpts: *str,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathkey := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathkey.PathName, os.ModePerm); err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	io.Copy(buf, r)

	filenameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameBytes[:])
	pathAndFilename := pathkey.PathName + "/" + filename

	f, err := os.Create(pathAndFilename)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}

	log.Printf("Written (%d) bytes to disk: %s", n, pathAndFilename)

	return nil
}
