package sivafuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type gitExample struct {
	Example  string
	Ok       bool
	PathType string
	Ref      string
	Path     string
}

var gitExamples = []gitExample{
	{
		Example:  "/this/is/a/path",
		Ok:       false,
		PathType: "",
		Ref:      "",
		Path:     "",
	},
	{"/_commit_/76683487299ab8", true, "commit", "76683487299ab8", ""},
	{"/_tag_/v0.6.5/some/path", true, "tag", "v0.6.5", "some/path"},
	{"/_branch_/master/other/path", true, "branch", "master", "other/path"},
	{"/_branch/master/other/path", false, "", "", ""},
}

func TestGetGitPath(t *testing.T) {
	for _, e := range gitExamples {
		ok, pathType, ref, path := getGitPath(e.Example)

		got := gitExample{
			Example:  e.Example,
			Ok:       ok,
			PathType: pathType,
			Ref:      ref,
			Path:     path,
		}

		assert.Equal(t, e, got)
	}
}

type commitExample struct {
	Example  string
	Ok       bool
	PathType string
	Path     string
}

var commitExamples = []commitExample{
	{
		Example:  "/this/is/a/path",
		Ok:       false,
		PathType: "",
		Path:     "",
	},
	{"message", true, "message", ""},
	{"/message", true, "message", ""},
	{"parent", true, "parent", ""},
	{"parent/0", true, "parent", "0"},
	{"parent/100", true, "parent", "100"},
	{"parent/100/12", false, "", ""},
	{"parent/test", false, "", ""},
	{"tree", true, "tree", ""},
	{"tree/src", true, "tree", "src"},
	{"tree/src/a", true, "tree", "src/a"},
	{"", true, "root", ""},
	{"08739d56c85059420/a42cbb342ccf4e68", false, "", ""},
}

func TestGetCommitPath(t *testing.T) {
	for _, e := range commitExamples {
		ok, pathType, path := getCommitPath(e.Example)

		got := commitExample{
			Example:  e.Example,
			Ok:       ok,
			PathType: pathType,
			Path:     path,
		}

		assert.Equal(t, e, got)
	}
}
