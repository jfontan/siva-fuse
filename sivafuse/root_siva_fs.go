package sivafuse

import (
	"io"
	"os"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"gopkg.in/src-d/go-billy-siva.v4"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

// RootSivaFS holds filesystem where siva files are stored
type RootSivaFS struct {
	path string
	pathfs.FileSystem
	FS     billy.Filesystem
	SivaFS sivafs.SivaFS
}

// NewRootSivaFs creates a new RootSivaFS from a path
func NewRootSivaFs(path string) *RootSivaFS {
	return &RootSivaFS{
		path:       path,
		FileSystem: pathfs.NewDefaultFileSystem(),
	}
}

func (r *RootSivaFS) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	ok, fsPath, sivaPath := getSivaPath(name)

	if ok {
		isGit, pType, ref, refPath := getGitPath(sivaPath)

		if isGit {
			siva, err := r.newSivaFS(fsPath)
			if err != nil {
				return "", fuse.ENOENT
			}

			git, err := GitOpen(siva)
			if err != nil {
				return "", fuse.ENOENT
			}

			link, err := git.Readlink(pType, ref, refPath)
			if err != nil {
				return "", fuse.ENOENT
			}

			return link, fuse.OK
		}
	}

	// return "lalalal", fuse.OK
	return "", fuse.ENOENT
}

// GetAttr returns file attributes
func (r *RootSivaFS) GetAttr(
	name string,
	context *fuse.Context,
) (*fuse.Attr, fuse.Status) {
	ok, fsPath, sivaPath := getSivaPath(name)

	var file os.FileInfo
	var err error

	if ok {
		isGit, pType, ref, refPath := getGitPath(sivaPath)

		if isGit && ref == "" {
			file = getTypeFileInfo(pType)
		} else {
			var siva sivafs.SivaFS
			siva, err = r.newSivaFS(fsPath)
			if err != nil {
				return nil, fuse.ENOENT
			}

			if isGit {
				var git *GitRepo
				git, err = GitOpen(siva)
				if err != nil {
					return nil, fuse.ENOENT
				}

				file, err = git.GetAttr(pType, ref, refPath)
			} else {
				file, err = siva.Stat(sivaPath)
				if file == nil {
					return nil, fuse.ENOENT
				}
			}
		}
	} else {
		file, err = r.FS.Stat(fsPath)
	}

	if err != nil || file == nil {
		return &fuse.Attr{}, fuse.ENOENT
	}

	var mode uint32
	if file.IsDir() {
		mode = 0500 | fuse.S_IFDIR
	} else {
		mode = 0400 | fuse.S_IFREG
	}

	if file.Mode()&syscall.S_IFLNK != 0 {
		mode |= syscall.S_IFLNK
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

// OpenDir returns the list of files in a given directory
func (r *RootSivaFS) OpenDir(
	name string,
	context *fuse.Context,
) (stream []fuse.DirEntry, code fuse.Status) {
	ok, fsPath, sivaPath := getSivaPath(name)

	var dir []os.FileInfo
	var err error

	if ok {
		var siva sivafs.SivaFS
		siva, err = r.newSivaFS(fsPath)
		if err != nil {
			return nil, fuse.ENOENT
		}

		isGit, pType, ref, refPath := getGitPath(sivaPath)

		if isGit {
			var git *GitRepo

			git, err = GitOpen(siva)
			if err != nil {
				return nil, fuse.ENOENT
			}

			dir, err = git.List(pType, ref, refPath)
			if err != nil {
				return nil, fuse.ENOENT
			}
		} else {
			dir, err = siva.ReadDir("/" + sivaPath)

			if err == nil && sivaPath == "" {
				dir = append(dir, getPathTypesFileInfo()...)
			}
		}
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

// Read fills a buffer with bytes from a file
func (s *billyFile) Read(
	dest []byte,
	off int64,
) (fuse.ReadResult, fuse.Status) {
	_, err := s.file.Seek(off, io.SeekStart)
	if err != nil {
		return nil, fuse.EINVAL
	}

	n, err := s.file.Read(dest)
	if err != nil {
		return nil, fuse.EINVAL
	}

	return fuse.ReadResultData(dest[:n]), fuse.OK
}

// Open retrieves a nodefs.File struct from a path
func (r *RootSivaFS) Open(
	name string,
	flags uint32,
	context *fuse.Context,
) (nodefs.File, fuse.Status) {
	ok, fsPath, sivaPath := getSivaPath(name)

	var f billy.File
	var err error

	if ok {
		var siva sivafs.SivaFS
		siva, err = r.newSivaFS(fsPath)
		if err != nil {
			return nil, fuse.ENOENT
		}

		isGit, pType, ref, path := getGitPath(sivaPath)

		if isGit {
			var git *GitRepo
			git, err = GitOpen(siva)

			if err == nil {
				var fuseFile nodefs.File
				fuseFile, err = git.Open(pType, ref, path)
				if err == nil {
					return fuseFile, fuse.OK
				}
			}
		} else {
			f, err = siva.OpenFile("/"+sivaPath, os.O_RDONLY, 0400)
		}
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
