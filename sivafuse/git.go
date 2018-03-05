package sivafuse

import (
	"io"
	"os"

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
	case "commit":
		return r.StatCommit(ref, path)

	default:
		return nil, os.ErrNotExist
	}
}

// List gets a FileInfo array of objects
func (r *GitRepo) List(pType, ref, path string) ([]os.FileInfo, error) {
	switch pType {
	case "commit":
		return r.ListCommits()

	default:
		return nil, os.ErrNotExist
	}
}

// StatCommit returns a FileInfo of the provided reference and path
func (r *GitRepo) StatCommit(ref, path string) (os.FileInfo, error) {
	if path != "" {
		return nil, os.ErrNotExist
	}

	commit, err := r.Repo.CommitObject(plumbing.NewHash(ref))
	if err != nil {
		return nil, err
	}

	return commitInfo(commit), nil
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

func commitInfo(commit *object.Commit) os.FileInfo {
	name := commit.Hash.String()
	text := commit.String()

	return NewFileInfo(name, int64(len(text)), false)
}
