package mfs

import (
	"os"
	"testing"
)

func TestCreateMfs(t *testing.T) {
	fs := NewFilesystem()

	if fs == nil {
		t.Error("Expected to create a file system")
	}

	if fs.root == nil || fs.root.name != "/" {
		t.Errorf("Root file was not created correctly")
	}
}

func TestMkdir(t *testing.T) {
	fs := NewFilesystem()

	paths := []string{"/tmp", "/tmp/foo", "/tmp/bar", "/etc"}
	for _, path := range paths {
		if err := fs.Mkdir(path, os.ModePerm); err != nil {
			t.Errorf("Mkdir %s: %s", path, err)
		}
	}

	for _, path := range paths {
		if _, stat, _ := fs.stat(path); stat == nil || !stat.IsDir() {
			t.Errorf("Directory %s was not created", path)
		}
	}

	for _, path := range []string{
		"/home/abiee", // Can't create a subdirectory if parent directory does not exists
		"/tmp",        // Not possible to create the same path twice
	} {
		if err := fs.Mkdir(path, os.ModePerm); err == nil {
			t.Errorf("Mkdir %q: no error", path)
		}
	}
}

func TestCreate(t *testing.T) {
	fs := NewFilesystem()

	filename := "/foo.txt"
	file, err := fs.Create(filename)
	if err != nil {
		t.Errorf("Create %q: %s", filename, err)
	}
	if file.Name() != filename {
		t.Errorf("Create %q, actual name %q while expected %q", filename, file.Name(), filename)
	}
	if _, stat, _ := fs.stat(filename); stat != nil {
		if stat.IsDir() {
			t.Errorf("Create %q, should not be a directory", filename)
		}
		if stat.mode != 0666 {
			t.Errorf("Create %q, wrong permissions 0%o while expected 0%o", filename, stat.mode, 0666)
		}
	} else {
		t.Errorf("File %s was not created", filename)
	}
	file.Close()

	// Overwrite file
	if file, err = fs.Create(filename); err != nil {
		t.Errorf("Create %q: %s", filename, err)
	}
	file.Close()

	// Should be able to write
	file, err = fs.Create(filename)
	if err != nil {
		t.Errorf("Create %q: %s", filename, err)
	}
	if _, err := file.Write([]byte("foo")); err != nil {
		t.Errorf("Unexpected error writing on file %q: %s", filename, err)
	}
	file.Close()

	fs.Mkdir("/tmp", os.ModePerm)
	for _, filename := range []string{
		"/tmp",         // Overwrite a directory
		"/etc/foo.txt", // Can't create a file in an inexistent directory
	} {
		if _, err = fs.Create(filename); err == nil {
			t.Errorf("Create %q: no error", filename)
		}
	}
}

func TestOpen(t *testing.T) {
	fs := NewFilesystem()
	filename := "/foo.txt"
	file, err := fs.Create(filename)
	if err != nil {
		t.Fatalf("Open: failed to create test file %q: %s", filename, err)
	}
	file.Close()

	file, err = fs.Open(filename)
	if err != nil {
		t.Errorf("Open %q: %s", filename, err)
	}
	if file.Name() != filename {
		t.Errorf("Open %q, actual name %q while expected %q", filename, file.Name(), filename)
	}
	file.Close()

	// Should not be able to write
	file, err = fs.Open(filename)
	if _, err := file.Write([]byte("error")); err == nil {
		t.Errorf("File %q should not be able to write: no error", filename)
	}
	file.Close()

	fs.Mkdir("/tmp", os.ModePerm)
	for _, filename := range []string{
		"/etc",         // Oopen a directory
		"/bar.txt",     // Inexistent file
		"/tmp/foo.txt", // Open file in inexistent directory
	} {
		if _, err = fs.Open(filename); err == nil {
			t.Errorf("Open %q: %s", filename, err)
		}
	}
}

func TestOpenFile(t *testing.T) {
	fs := NewFilesystem()
	filename := "/foo.txt"
	file, err := fs.Create(filename)
	if err != nil {
		t.Fatalf("OpenFile: failed to create test file %q: %s", filename, err)
	}
	file.Close()

	file, err = fs.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		t.Errorf("OpenFile(O_WRONLY) %q: %s", filename, err)
	}
	if _, err = file.Write([]byte(lorem)); err != nil {
		t.Errorf("Write(O_WRONLY) %q: %s", filename, err)
	}
	if _, err = file.Read(make([]byte, 10)); err != ErrWriteOnly {
		t.Errorf("Read(O_WRONLY) %q: %s", filename, err)
	}
	file.Close()

	file, err = fs.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		t.Errorf("OpenFile(O_RDONLY) %q: %s", filename, err)
	}
	if _, err = file.Read(make([]byte, 5)); err != nil {
		t.Errorf("Read(O_RDONLY) %q: %s", filename, err)
	}
	if _, err = file.Write([]byte("error")); err != ErrReadOnly {
		t.Errorf("Write(O_RDONLY) %q: %s", filename, err)
	}
	file.Close()

	file, err = fs.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		t.Errorf("OpenFile(O_RDWR) %q: %s", filename, err)
	}
	if _, err = file.Read(make([]byte, 5)); err != nil {
		t.Errorf("OpenFile(O_RDWR) %q: %s", filename, err)
	}
	if _, err = file.Write([]byte(lorem)); err != nil {
		t.Errorf("Write(O_RDONLY) %q: %s", filename, err)
	}
	file.Close()
}

func TestOpenFileAppend(t *testing.T) {
	fs := NewFilesystem()
	filename := "/foo.txt"
	file, err := fs.Create(filename)
	if err != nil {
		t.Fatalf("OpenFile: failed to create test file %q: %s", filename, err)
	}
	if _, err = file.Write([]byte(lorem)); err != nil {
		t.Fatalf("OpenFile: failed to write on test file %q: %s", filename, err)
	}
	file.Close()

	data := make([]byte, len(lorem)+len(hello))
	file, err = fs.OpenFile(filename, os.O_APPEND|os.O_RDWR, 0)
	if err != nil {
		t.Errorf("OpenFile(O_APPEND|O_RDWR) %q: %s", filename, err)
	}
	if _, err = file.Write([]byte(hello)); err != nil {
		t.Errorf("Write(O_APPEND|O_RDWR) %q: %s", filename, err)
	}
	if _, err = file.Seek(0, os.SEEK_SET); err != nil {
		t.Fatalf("Seek(O_APPEND|O_RDWR) %q: %s", filename, err)
	}
	if _, err = file.Read(data); err != nil {
		t.Errorf("Read(O_APPEND|O_RDWR) %q: %s", filename, err)
	}
	if string(data) != lorem+hello {
		t.Errorf("Wrong O_APPEND operation, file contains %q", data)
	}
	file.Close()
}

func TestOpenFileCreate(t *testing.T) {
	var content []byte

	fs := NewFilesystem()
	filename := "/foo.txt"

	content = make([]byte, 12)
	file, err := fs.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Errorf("OpenFile(os.O_CREATE|os.O_WRONLY) %q: %s", filename, err)
	}
	if _, err := file.Write([]byte("Hello world!")); err != nil {
		t.Errorf("Write(os.O_CREATE|os.O_WRONLY) %q: %s", filename, err)
	}
	if _, entry, _ := fs.stat(filename); entry == nil {
		t.Errorf("FIle %q was not created", filename)
	}
	file.Close()
	file, err = fs.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("OpenFile %q: %s", filename, err)
	}
	if _, err = file.Read(content); err != nil {
		t.Errorf("Read test file %q: %s", filename, err)
	}
	if string(content) != "Hello world!" {
		t.Errorf("File expected to contain \"Hello world!\" but got %q", content)
	}
	file.Close()

	// Overwrite an existent file with O_TRUNC
	content = make([]byte, 10)
	file, err = fs.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Errorf("OpenFile(os.O_CREATE|os.O_TRUNC|os.O_WRONLY) %q: %s", filename, err)
	}
	if _, err = file.Write([]byte("Bye world!")); err != nil {
		t.Errorf("Write(os.O_CREATE|os.O_TRUNC|os.O_WRONLY) %q: %s", filename, err)
	}
	file.Close()
	file, err = fs.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("OpenFile %q: %s", filename, err)
	}
	if _, err = file.Read(content); err != nil {
		t.Errorf("Read test file %q: %s", filename, err)
	}
	if string(content) != "Bye world!" {
		t.Errorf("File expected to contain \"Bye world!\" but got %q", content)
	}
	file.Close()

	if _, err = fs.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0); err == nil {
		t.Errorf("OpenFile(os.O_CREATE|os.O_WRONLY) overwriting %q: no error", filename)
	}
}

func TestStat(t *testing.T) {
	fs := NewFilesystem()
	fs.Mkdir("/tmp", os.ModePerm)
	filename := "/foo.txt"
	file, err := fs.Create(filename)
	if err != nil {
		t.Fatalf("OpenFile: failed to create test file %q: %s", filename, err)
	}
	if _, err = file.Write([]byte(lorem)); err != nil {
		t.Fatalf("OpenFile: failed to write on test file %q: %s", filename, err)
	}
	file.Close()

	for _, test := range []struct {
		Path     string
		Filename string
		Size     int64
		IsDir    bool
	}{
		{"/tmp", "tmp", 0, true},
		{"/foo.txt", "foo.txt", int64(len(lorem)), false},
	} {
		info, err := fs.Stat(test.Path)
		if err != nil {
			t.Errorf("Stat %q: %s", test.Path, err)
		}
		if info.Name() != test.Filename {
			t.Errorf("Expected filename %q but got %q", test.Filename, info.Name())
		}
		if info.IsDir() != test.IsDir {
			t.Errorf("Expected %q to IsDir=%t", test.Filename, test.IsDir)
		}
		if info.Size() != test.Size {
			t.Errorf("Expected %q to have a size of %d but got %d", test.Filename, test.Size, info.Size())
		}
	}

	for _, filename := range []string{"/bar.txt", "/etc/foo.txt"} {
		if _, err := fs.Stat(filename); err == nil {
			t.Errorf("Stat %q: no error", filename)
		}
	}
}
