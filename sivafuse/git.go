package sivafuse

import (
	"io"
	"os"

	"gopkg.in/src-d/go-billy-siva.v4"
	"gopkg.in/src-d/go-git.v4"
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

// DirCommits returns a FileInfo array of commit hashes
func (r *GitRepo) DirCommits() ([]os.FileInfo, error) {
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

		name := c.Hash.String()
		text := c.String()

		files = append(files, NewFileInfo(name, int64(len(text)), false))
	}

	return files, nil
}
