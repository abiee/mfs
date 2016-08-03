Memory filesystem (mfs)
===================
mfs allows you to create a filesystem in memory for testing proposes. The reason why this library exists is to make a test only dependency.

Install
---
You can install mfs with `go get`.

    go get github.com/abiee/mfs

Available functions
---

 - `Create(name string) (*File, error)`
 - `Open(name string) (*File, error)`
 - `OpenFile(name string, flag int, perm FileMode) (*File, error)`
 - `Mkdir(name string, perm FileMode) error`
 - `Stat(name string) (FileInfo, error)`

Usage
---
Here is an example how it works.

    fs := NewFs()

    file, err := fs.Create("/home/abiee/hello.txt")
    if err != nil {
	    log.Fatal(err)
    }

	if _, err := file.Write([]byte("Hello world")); err != nil {
	    log.Fatal(err)
	}

	file.Close()

How to use it
---
You will need to create an interface to a filesystem.

    struct File interface {
        io.Writer
		io.ReadSeeker
		Name() string
		Close() error
    }

    struct Filesystem interface {
	    Create(name string) (File, error)
	    Open(name string) (File, error)
	    OpenFile(name string, flag int, perm FileMode) (File, error)
	    Mkdir(name string, perm os.FileMode) error
	    Stat(name string) (os.FileInfo, error)
    }

These interface functions signatures are the same that found on the `os` package, so that you can create an straightforward implementation.

    struct OsFilesystem struct {}
    
    func (o OsFilesystem) Create(name string) (*File, error) {
        return os.Create(name)
    }

    func (o OsFilesystem) Open(name string) (*File, error) {
        return os.Open(name)
    }

    // ...

Then you can create a mfs implementation.

    struct FakeFilesystem struct {
        Filesystem
    }
    
    func (fs FakeFilesystem) Create(name string) (*File, error) {
        return fs.Create(name)
    }

    // ...

How to make tests
---

Imagine you have a function like that in your code.

    func DoSomething(fs Filesystem) err {
	    file, err := fs.Create("/foo.txt")
	    if err != nil {
		    return err
	    }
	    // do something with file
    }

In your tests you can inject the `FakeFilesystem`.

    func TestDoSomething(t *testing.T) {
	    fs := FakeFilesystem{mfs.NewFilesystem()}
	    if err := DoSomething(fs); err != nil {
		    t.Fatalf("Unexpected error: %s", err)
	    }
	    if stat, _ := fs.Stat("/foo.txt"); stat == nil {
		    t.Errorf("File /foo.txt was not created")
	    }
    }
