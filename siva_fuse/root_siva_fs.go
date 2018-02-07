package sivaFuse

import (
	"io"
	"os"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"gopkg.in/src-d/go-billy-siva.v4"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

type RootSivaFS struct {
	path string
	pathfs.FileSystem
	FS     billy.Filesystem
	SivaFS sivafs.SivaFS
}

func NewRootSivaFs(path string) *RootSivaFS {
	return &RootSivaFS{
		path:       path,
		FileSystem: pathfs.NewDefaultFileSystem(),
	}
}

func (r *RootSivaFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	ok, fsPath, sivaPath := getSivaPath(name)

	// TODO: why can not stat directories from siva?

	var file os.FileInfo
	var err error

	if ok {
		var siva sivafs.SivaFS
		siva, err = r.newSivaFS(fsPath)
		if err != nil {
			return nil, fuse.ENOENT
		}

		file, err = siva.Stat(sivaPath)
		if file == nil {
			return nil, fuse.ENOENT
		}
	} else {
		file, err = r.FS.Stat(fsPath)
	}

	if err != nil {
		return &fuse.Attr{}, fuse.ENOENT
	}

	if file == nil {
		return &fuse.Attr{}, fuse.ENOENT
	}

	mode := uint32(0)
	if file.IsDir() {
		mode = 0500 | fuse.S_IFDIR
	} else {
		mode = 0400 | fuse.S_IFREG
	}

	a := fuse.Attr{
		Owner: *fuse.CurrentOwner(),
		Mode:  mode,
		Size:  uint64(file.Size()),
	}

	return &a, fuse.OK
}

func (r *RootSivaFS) newSivaFS(fsPath string) (sivafs.SivaFS, error) {
	return sivafs.NewFilesystem(r.FS, fsPath, memfs.New())
}

func (r *RootSivaFS) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	ok, fsPath, sivaPath := getSivaPath(name)

	var dir []os.FileInfo
	var err error

	if ok {
		var siva sivafs.SivaFS
		siva, err = r.newSivaFS(fsPath)
		if err != nil {
			return nil, fuse.ENOENT
		}

		dir, err = siva.ReadDir("/" + sivaPath)
	} else {
		dir, err = r.FS.ReadDir("/" + name)
	}

	if err != nil {
		return nil, fuse.ENOENT
	}

	d := make([]fuse.DirEntry, 0)
	for _, file := range dir {
		f := fuse.DirEntry{
			Name: file.Name(),
			Mode: 0500,
		}
		d = append(d, f)
	}
	return d, fuse.OK
}

type billyFile struct {
	nodefs.File
	file billy.File
}

func (s *billyFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	s.file.Seek(off, io.SeekStart)

	n, err := s.file.Read(dest)
	if err != nil {
		return nil, fuse.EINVAL
	}

	return fuse.ReadResultData(dest[:n]), fuse.OK
}

func (r *RootSivaFS) Open(name string, flags uint32,
	context *fuse.Context) (nodefs.File, fuse.Status) {

	ok, fsPath, sivaPath := getSivaPath(name)

	var f billy.File
	var err error

	if ok {
		var siva sivafs.SivaFS
		siva, err = r.newSivaFS(fsPath)
		if err != nil {
			return nil, fuse.ENOENT
		}

		f, err = siva.OpenFile("/"+sivaPath, os.O_RDONLY, 0400)
	} else {
		f, err = r.FS.OpenFile(fsPath, os.O_RDONLY, 0400)
	}

	if err != nil {
		return nil, fuse.ENOSYS
	}

	return &billyFile{
		file: f,
		File: nodefs.NewDefaultFile(),
	}, fuse.OK
}
