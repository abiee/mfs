package mfs

import (
	"io"
	"os"
	"testing"
)

const (
	hello = "Hello world!"
	bye   = "Bye world"
	lorem = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
)

func newFile(name, content string) *memoryFile {
	data := []byte(content)

	return &memoryFile{
		name: name,
		data: &data,
	}
}

func TestWrite(t *testing.T) {
	file := newFile("foo.txt", "")
	defer file.Close()

	if n, err := file.Write([]byte(hello)); err != nil || n != len(hello) {
		t.Errorf("Write: (%d bytes) %s", n, err)
	}
	if string((*file.data)[:len(hello)]) != hello {
		t.Errorf("Wrong file content %q expected %q", *file.data, hello)
	}
	if n, err := file.Write([]byte(bye)); err != nil || n != len(bye) {
		t.Errorf("Write: (%d bytes) %s", n, err)
	}
	if string((*file.data)[:len(hello)+len(bye)]) != hello+bye {
		t.Errorf("Wrong file content %q expected %q", *file.data, hello+bye)
	}
}

func TestRead(t *testing.T) {
	var data []byte

	file := newFile("foo.txt", hello)
	defer file.Close()

	testCases := []struct {
		Size     int
		Expected string
	}{
		{Size: 5, Expected: hello[:5]},
		{Size: 5, Expected: hello[5:10]},
		{Size: 2, Expected: hello[10:12]},
	}

	for _, test := range testCases {
		data = make([]byte, test.Size)
		if n, err := file.Read(data); err != nil || n != len(data) {
			t.Errorf("Read: (%d bytes) %s", n, err)
		}
		if string(data) != test.Expected {
			t.Errorf("Readed %q expected %q", data, test.Expected)
		}
	}

	data = make([]byte, 1)
	if n, err := file.Read(data); err != io.EOF || n != 0 {
		t.Errorf("Expected to be EOF: %d %s", n, err)
	}
}

func TestSeek(t *testing.T) {
	data := make([]byte, 5)
	file := newFile("foo.txt", lorem)
	defer file.Close()

	testCases := []struct {
		Offset   int64
		Whence   int
		Expected string
	}{
		{10, os.SEEK_SET, lorem[10:15]},
		{5, os.SEEK_CUR, lorem[20:25]},
		{-5, os.SEEK_END, lorem[len(lorem)-5:]},
	}

	for _, test := range testCases {
		if offset, err := file.Seek(test.Offset, test.Whence); err != nil {
			t.Errorf("Seek(%d@%d): %s", test.Whence, offset, err)
		}
		if _, err := file.Read(data); err != nil {
			t.Errorf("Read file: %s", err)
		}
		if test.Expected != string(data) {
			t.Errorf("Expected to read %q got %q", test.Expected, data)
		}
	}

	for _, test := range []struct {
		Offset int64
		Whence int
		Error  error
	}{
		{-1, os.SEEK_SET, ErrNegativeSeek},
		{int64(len(lorem) + 1), os.SEEK_SET, ErrTooFar},
		{1, 100, ErrInvalidWhence},
	} {
		if n, err := file.Seek(test.Offset, test.Whence); err != test.Error || n != 0 {
			t.Errorf("Seek %d: %s", n, err)
		}
	}
}
