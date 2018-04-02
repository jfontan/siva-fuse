package sivafuse

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type refInfo struct {
	name string
	dir  bool
	link string
}

func newRefInfo(name string, dir bool, link string) *refInfo {
	return &refInfo{
		name: name,
		dir:  dir,
		link: link,
	}
}

func newRefDir(name string) *refInfo {
	return newRefInfo(name, true, "")
}

func newRefLink(name, link string) *refInfo {
	return newRefInfo(name, false, link)
}

func statBranch(repo *git.Repository, ref string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

func statTag(repo *git.Repository, ref string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

func statRef(repo *git.Repository, ref string) (os.FileInfo, error) {
	branches, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	for {
		b, err := branches.Next()
		if err == io.EOF {
			return nil, os.ErrNotExist
		}

		if err != nil {
			return nil, err
		}

		if b.Name().Short() == ref {
			return NewLinkInfo(ref), nil
		}
	}
}

func listBranch(repo *git.Repository, ref string) ([]os.FileInfo, error) {
	branches, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	return listRef(branches, ref)
}

func listTag(repo *git.Repository, ref string) ([]os.FileInfo, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	return listRef(tags, ref)
}

func splitRef(ref string, level int) (string, []string) {
	split := strings.Split(ref, "/")

	if level == 0 {
		return "", split
	}

	if len(split) <= level {
		return ref, nil
	}

	base := strings.Join(split[:level], "/")
	return base, split[level:]
}

func getRefs(
	iter storer.ReferenceIter,
	ref string,
) ([]*refInfo, error) {
	refs := make([]*refInfo, 0, 2)

	_, split := splitRef(ref, 0)
	level := len(split)
	lastName := split[len(split)-1]

	// searching root references
	if ref == "" {
		level = 0
		lastName = ""
	}

	found := make(map[string]struct{}, 10)

	for {
		t, err := iter.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		name := t.Name().Short()
		base, path := splitRef(name, level)

		// _tag_/fix/bug ../../_commit_/aabbccddee
		_, parts := splitRef(name, 0)
		dots := strings.Repeat("../", len(parts))
		prefix := filepath.Join(dots, "_commit_")
		linkPath := filepath.Join(prefix, t.Hash().String())

		// We've found a reference with the exact same name
		if ref == name {
			return []*refInfo{
				newRefLink(lastName, linkPath),
			}, nil
		}

		// is not part of the reference tree
		if ref != base {
			continue
		}

		if len(path) == 0 {
			continue
		}

		// Do not add files more than once. It can happen for nested paths:
		//
		// fix/1
		// fix/2
		// fix/2
		if _, ok := found[path[0]]; ok {
			continue
		}
		found[path[0]] = struct{}{}

		if len(path) == 1 {
			refs = append(refs, newRefLink(path[0], linkPath))
		} else {
			refs = append(refs, newRefDir(path[0]))
		}
	}

	return refs, nil
}

func listRef(iter storer.ReferenceIter, ref string) ([]os.FileInfo, error) {
	return nil, nil
}
