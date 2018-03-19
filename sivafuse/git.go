package sivafuse

import (
	"bytes"
	"io"
	"os"
	"strconv"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"gopkg.in/src-d/go-billy-siva.v4"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

// GitRepo contains a git repository and helper functions
type GitRepo struct {
	Repo *git.Repository
}

// GitOpen creates a new GitRepo from a sivafs
func GitOpen(sivaFS sivafs.SivaFS) (*GitRepo, error) {
	storage, err := filesystem.NewStorage(sivaFS)

	if err != nil {
		return nil, err
	}

	g, err := git.Open(storage, nil)
	if err != nil {
		return nil, err
	}

	return &GitRepo{Repo: g}, err
}

func (r *GitRepo) GetAttr(pType, ref, path string) (os.FileInfo, error) {
	switch pType {
	case tCommit:
		return r.StatCommit(ref, path)

	default:
		return nil, os.ErrNotExist
	}
}

// List gets a FileInfo array of objects
func (r *GitRepo) List(pType, ref, path string) ([]os.FileInfo, error) {
	switch pType {
	case tCommit:
		return r.ListCommit(ref, path)

	default:
		return nil, os.ErrNotExist
	}
}

// StatCommit returns a FileInfo of the provided reference and path
func (r *GitRepo) StatCommit(ref, p string) (os.FileInfo, error) {
	ok, pathType, path := getCommitPath(p)
	if !ok {
		return nil, os.ErrNotExist
	}

	return r.statCommitFile(ref, pathType, path)
}

func (r *GitRepo) ListCommit(ref, path string) ([]os.FileInfo, error) {
	if ref == "" && path == "" {
		return r.ListCommits()
	}

	ok, pathType, p := getCommitPath(path)
	if !ok {
		return nil, os.ErrNotExist
	}

	return r.listCommitDirectory(ref, pathType, p)
}

func (r *GitRepo) statCommitFile(ref, pType, path string) (os.FileInfo, error) {
	commit, err := r.Repo.CommitObject(plumbing.NewHash(ref))
	if err != nil {
		return nil, err
	}

	switch pType {
	case tRoot:
		return NewFileInfo(ref, 0, true), nil
	case tMessage:
		return NewFileInfo(tMessage, int64(len(commit.String())), false), nil
	case tParent:
		return NewFileInfo(tParent, 0, true), nil
	case tTree:
		return r.statTreeFile(ref, path)
	}

	return nil, os.ErrNotExist
}

func (r *GitRepo) statTreeFile(ref, path string) (os.FileInfo, error) {
	if path == "" {
		return NewFileInfo(tTree, 0, true), nil
	}

	commit, err := r.Repo.CommitObject(plumbing.NewHash(ref))
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	f, err := tree.FindEntry(path)
	if err != nil {
		return nil, err
	}

	size := int64(0)
	if f.Mode.IsFile() {
		file, err := tree.TreeEntryFile(f)
		if err != nil {
			return nil, err
		}

		size = file.Size
	}

	return NewFileInfo(f.Name, size, !f.Mode.IsFile()), nil
}

func (r *GitRepo) listCommitDirectory(ref, ptype, path string) ([]os.FileInfo, error) {
	switch ptype {
	case tRoot:
		files := make([]os.FileInfo, 3)

		for i, file := range []string{tMessage, tParent, tTree} {
			f, err := r.statCommitFile(ref, file, "")
			if err != nil {
				return nil, err
			}

			files[i] = f
		}

		return files, nil

	case tParent:
		if path != "" {
			return nil, os.ErrNotExist
		}

		return r.listParents(ref)

	case tTree:
		return r.listTree(ref, path)
	}

	return nil, os.ErrNotExist
}

// ListCommits returns a FileInfo array of commit hashes
func (r *GitRepo) ListCommits() ([]os.FileInfo, error) {
	commits, err := r.Repo.CommitObjects()
	if err != nil {
		return nil, err
	}

	defer commits.Close()

	files := make([]os.FileInfo, 0, 16)

	for {
		c, err := commits.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}

			break
		}

		files = append(files, commitInfo(c))
	}

	return files, nil
}

func (r *GitRepo) listParents(ref string) ([]os.FileInfo, error) {
	commit, err := r.Repo.CommitObject(plumbing.NewHash(ref))
	if err != nil {
		return nil, err
	}

	parents := commit.Parents()
	if err != nil {
		return nil, err
	}

	files := make([]os.FileInfo, 0, 2)

	pos := 0
	for {
		_, err := parents.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		files = append(files, NewFileInfo(strconv.Itoa(pos), 0, false))
		pos++
	}

	return files, nil
}

func (r *GitRepo) listTree(ref, path string) ([]os.FileInfo, error) {
	commit, err := r.Repo.CommitObject(plumbing.NewHash(ref))
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	if path != "" {
		tree, err = tree.Tree(path)
		if err != nil {
			return nil, err
		}
	}

	files := make([]os.FileInfo, len(tree.Entries))

	for i, f := range tree.Entries {
		files[i] = NewFileInfo(f.Name, 0, !f.Mode.IsFile())
	}

	return files, nil
}

func commitInfo(commit *object.Commit) os.FileInfo {
	name := commit.Hash.String()
	return NewFileInfo(name, 0, true)
}

func (r *GitRepo) Open(pType, ref, path string) (nodefs.File, error) {
	switch pType {
	case tCommit:
		return r.OpenCommit(ref, path)

	default:
		return nil, os.ErrNotExist
	}
}

func (r *GitRepo) OpenCommit(ref, path string) (nodefs.File, error) {
	ok, pType, _ := getCommitPath(path)
	if !ok {
		return nil, os.ErrNotExist
	}

	commit, err := r.Repo.CommitObject(plumbing.NewHash(ref))
	if err != nil {
		return nil, err
	}

	switch pType {
	case tMessage:
		reader := bytes.NewBufferString(commit.String())
		closer := &readCloser{reader}
		file := NewFuseFile(closer)
		return file, nil
	}

	return nil, os.ErrNotExist
}

type readCloser struct {
	io.Reader
}

func (readCloser) Close() error {
	return nil
}

type fuseFile struct {
	nodefs.File
	reader io.ReadCloser
}

func NewFuseFile(read io.ReadCloser) *fuseFile {
	return &fuseFile{
		File:   nodefs.NewDefaultFile(),
		reader: read,
	}
}

// Read fills a buffer with bytes from a reader
func (f *fuseFile) Read(
	dest []byte,
	off int64,
) (fuse.ReadResult, fuse.Status) {
	// Skip offset bytes
	if off != 0 {
		buf := make([]byte, off)
		_, err := f.reader.Read(buf)
		if err != nil {
			return nil, fuse.EINVAL
		}
	}

	n, err := f.reader.Read(dest)
	if err != nil {
		return nil, fuse.EINVAL
	}

	return fuse.ReadResultData(dest[:n]), fuse.OK
}

func (f *fuseFile) Close() error {
	return f.reader.Close()
}
