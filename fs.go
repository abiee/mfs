package mfs

import (
	"errors"
	"os"
	"path"
	"strings"
	"time"
)

const (
	PathSeparator     = "/"
	InitialDataSize   = 512
	CreateDefaultMode = 0666
)

var ErrDirectory = errors.New("File is a directory")

type MemoryFilesystem struct {
	root *fileInfo
}

func NewFilesystem() *MemoryFilesystem {
	root := &fileInfo{
		name:     "/",
		children: make(map[string]*fileInfo),
	}
	return &MemoryFilesystem{root}
}

func (fs *MemoryFilesystem) Mkdir(name string, perm os.FileMode) error {
	base := path.Base(name)

	parent, node, err := fs.stat(name)
	if err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}
	if node != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: os.ErrExist}
	}

	parent.children[base] = &fileInfo{
		name:     base,
		fullpath: name,
		mode:     perm,
		children: make(map[string]*fileInfo),
	}

	return nil
}

func (fs *MemoryFilesystem) Create(name string) (File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, CreateDefaultMode)
}

func (fs *MemoryFilesystem) Open(name string) (File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *MemoryFilesystem) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	base := path.Base(name)
	parent, node, err := fs.stat(name)

	if err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}

	if node != nil && node.IsDir() {
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrDirectory}
	}

	if hasFlag(flag, os.O_CREATE) {
		if node != nil && !hasFlag(flag, os.O_TRUNC) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrExist}
		}

		data := make([]byte, 0, InitialDataSize)
		node = &fileInfo{
			name:     base,
			fullpath: name,
			mode:     perm,
			data:     &data,
		}
		parent.children[base] = node
	}

	if node == nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	return node.openFile(flag)
}

func (fs *MemoryFilesystem) Stat(name string) (os.FileInfo, error) {
	_, node, err := fs.stat(name)

	if err != nil {
		return nil, &os.PathError{Op: "stat", Path: name, Err: err}
	}
	if node == nil {
		return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
	}

	return node, nil
}

func (fs *MemoryFilesystem) stat(name string) (*fileInfo, *fileInfo, error) {
	segments := strings.Split(name, PathSeparator)
	parent := fs.root

	for _, segment := range segments[1 : len(segments)-1] {
		if entry, ok := parent.children[segment]; ok && entry.IsDir() {
			parent = entry
		} else {
			return nil, nil, os.ErrNotExist
		}
	}

	segment := segments[len(segments)-1]
	if node, ok := parent.children[segment]; ok {
		return parent, node, nil
	}

	return parent, nil, nil
}

type fileInfo struct {
	name     string
	fullpath string
	mode     os.FileMode
	children map[string]*fileInfo
	data     *[]byte
}

func (f fileInfo) Name() string {
	return f.name
}

func (f fileInfo) Size() int64 {
	if f.IsDir() {
		return 0
	}

	return int64(len(*f.data))
}

func (f fileInfo) ModTime() time.Time {
	return time.Now()
}

func (f fileInfo) Mode() os.FileMode {
	return f.mode
}

func (f fileInfo) IsDir() bool {
	return f.children != nil
}

func (f fileInfo) Sys() interface{} {
	return nil
}

func (fi *fileInfo) openFile(flag int) (File, error) {
	file := &memoryFile{
		name: fi.fullpath,
		data: fi.data,
	}

	if hasFlag(flag, os.O_APPEND) {
		file.Seek(0, os.SEEK_END)
	}

	if hasFlag(flag, os.O_RDWR) {
		return file, nil
	} else if hasFlag(flag, os.O_WRONLY) {
		return &writeOnlyFile{file}, nil
	} else {
		return &readOnlyFile{file}, nil
	}
}

func hasFlag(flag, target int) bool {
	return flag&target == target
}
