package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	defaultFilePerm os.FileMode = 0666
	defaultPathPerm os.FileMode = 0777
)

// nopWriteCloser wraps an io.Writer and provides a no-op Close method to
// satisfy the io.WriteCloser interface.
type nopWriteCloser struct {
	io.Writer
}

func (wc *nopWriteCloser) Write(p []byte) (int, error) { return wc.Writer.Write(p) }
func (wc *nopWriteCloser) Close() error                { return nil }

// IsDir return true if the path is a directory
func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

// FileWalk type channel that represent interation trough a path
type FileWalk chan string

// Walk do the walk
func (f FileWalk) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		f <- path
	}
	return nil
}

// StringInSlice Return true if a is in the list slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// WriteReaderToFile will write an io.Reader to file dest
func WriteReaderToFile(r io.Reader, dest string) error {
	mode := os.O_WRONLY | os.O_CREATE | os.O_TRUNC // overwrite if exists
	f, err := os.OpenFile(dest, mode, defaultFilePerm)
	if err != nil {
		return fmt.Errorf("open file: %s", err)
	}

	wc := io.WriteCloser(&nopWriteCloser{f})

	if _, err := io.Copy(wc, r); err != nil {
		f.Close() // error deliberately ignored
		return fmt.Errorf("i/o copy: %s", err)
	}

	if err := wc.Close(); err != nil {
		f.Close() // error deliberately ignored
		return fmt.Errorf("compression close: %s", err)
	}

	if err := f.Sync(); err != nil {
		f.Close() // error deliberately ignored
		return fmt.Errorf("file sync: %s", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("file close: %s", err)
	}

	return nil
}

// GetByteHash will return hex sha264 of specified byte sequence
func GetByteHash(bt []byte) string {
	h := sha256.New()
	h.Write(bt)

	return fmt.Sprintf("%x", h.Sum(nil))
}

// GetStringHash will return string hash of specified string
func GetStringHash(plainString string) string {
	return GetByteHash([]byte(plainString))
}

// GetFileHash will return string hash of specified file
func GetFileHash(path string) string {
	slurp, _ := ioutil.ReadFile(path)
	return GetByteHash(slurp)
}
