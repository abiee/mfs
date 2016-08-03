package mfs

import (
	"errors"
	"io"
	"os"
)

var (
	ErrReadOnly      = errors.New("Read-only file")
	ErrWriteOnly     = errors.New("Write-only file")
	ErrTooLarge      = errors.New("File too large")
	ErrNegativeSeek  = errors.New("Negative seek offset")
	ErrInvalidWhence = errors.New("Invalid seek whence")
	ErrTooFar        = errors.New("Too far")
)

type File interface {
	io.Writer
	io.ReadSeeker
	io.Closer
	Name() string
}

type memoryFile struct {
	name   string
	offset int64
	data   *[]byte
}

func (f *memoryFile) Name() string {
	return f.name
}

func (f *memoryFile) Read(p []byte) (int, error) {
	if f.offset >= f.size() {
		return 0, io.EOF
	}

	n := copy(p, (*f.data)[f.offset:])
	f.offset += int64(n)

	return n, nil
}

func (f *memoryFile) Write(p []byte) (int, error) {
	n := len(p)

	if f.offset+int64(n) > f.size() {
		err := f.grow(n)
		if err != nil {
			return 0, err
		}
	}

	copy((*f.data)[f.offset:], p)
	f.offset += int64(n)

	return n, nil
}

func (f *memoryFile) size() int64 {
	return int64(len(*f.data))
}

func (f *memoryFile) grow(n int) error {
	m := len(*f.data)

	if m+n > cap(*f.data) {
		newCap := 2*cap(*f.data) + n

		data, err := makeSlice(newCap)
		if err != nil {
			return err
		}

		copy(data, *f.data)
		f.data = &data
	}

	*f.data = (*f.data)[0 : m+n]
	return nil
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) (slice []byte, err error) {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			slice = nil
			err = ErrTooLarge
			return
		}
	}()
	slice = make([]byte, n)
	return
}

func (f *memoryFile) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64

	switch whence {
	case os.SEEK_SET:
		newOffset = offset
	case os.SEEK_CUR:
		newOffset = f.offset + offset
	case os.SEEK_END:
		newOffset = f.size() + offset
	default:
		return 0, ErrInvalidWhence
	}

	if newOffset < 0 {
		return 0, ErrNegativeSeek
	}

	if newOffset > f.size() {
		return 0, ErrTooFar
	}

	f.offset = newOffset

	return f.offset, nil
}

func (f *memoryFile) Close() error {
	return nil
}

type readOnlyFile struct {
	File
}

func (f *readOnlyFile) Write(p []byte) (int, error) {
	return 0, ErrReadOnly
}

type writeOnlyFile struct {
	File
}

func (f *writeOnlyFile) Read(p []byte) (int, error) {
	return 0, ErrWriteOnly
}
