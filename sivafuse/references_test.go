package sivafuse

import (
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type mockRef struct {
	name string
	hash string
}

type mockRefIter struct {
	data []mockRef
	pos  int
}

func newMockRefIter(data []mockRef) storer.ReferenceIter {
	return &mockRefIter{
		data: data,
		pos:  0,
	}
}

func (m *mockRefIter) Next() (*plumbing.Reference, error) {
	if m.pos >= len(m.data) {
		return nil, io.EOF
	}

	d := m.data[m.pos]
	m.pos++

	return plumbing.NewReferenceFromStrings(d.name, d.hash), nil
}

func (m *mockRefIter) ForEach(func(*plumbing.Reference) error) error {
	panic("not implemented")
}

func (m *mockRefIter) Close() {
	panic("not implemented")
}

type refExample struct {
	Example  string
	Expected []*refInfo
}

func cp(level int, hash string) string {
	prefix := strings.Repeat("../", level)
	return filepath.Join(prefix, "_commit_", hash)
}

var mockBranches = []mockRef{
	// branches
	{"refs/heads/master", "f2d45df860b4da12d7658b7a083029a66217a90b"},
	{"refs/heads/fix/bug1", "a0b01b2d4a98a5bdb4ee7e3b5ad2b39330d28b8b"},
	{"refs/heads/fix/bug2", "26f052fc2c880ea3c14394aa7b9763f68668a26b"},
	{"refs/heads/fix/deep/branch", "f93ab58dff804d5f65a53cdd8f0945bf7662b4c2"},
}

var branchExamples = []refExample{
	{
		Example:  "this/is/a/path",
		Expected: []*refInfo{},
	},
	{"master", []*refInfo{
		newRefLink("master", cp(1, "f2d45df860b4da12d7658b7a083029a66217a90b"))},
	},
	{"", []*refInfo{
		newRefLink("master", cp(1, "f2d45df860b4da12d7658b7a083029a66217a90b")),
		newRefDir("fix")},
	},
	{"fix", []*refInfo{
		newRefLink("bug1", cp(2, "a0b01b2d4a98a5bdb4ee7e3b5ad2b39330d28b8b")),
		newRefLink("bug2", cp(2, "26f052fc2c880ea3c14394aa7b9763f68668a26b")),
		newRefDir("deep")},
	},
	{"fix/deep", []*refInfo{
		newRefLink("branch", cp(3, "f93ab58dff804d5f65a53cdd8f0945bf7662b4c2"))},
	},
}

func TestGetRefsBranch(t *testing.T) {
	require := require.New(t)

	for _, e := range branchExamples {
		iter := newMockRefIter(mockBranches)
		refs, err := getRefs(iter, e.Example)
		require.NoError(err)

		got := refExample{
			Example:  e.Example,
			Expected: refs,
		}

		require.Equal(e, got)
	}
}

var mockTags = []mockRef{
	{"refs/tags/v0.0.1", "5a57370fe617240f839a8a2e41f3b23c97f008fb"},
	{"refs/tags/v0.0.2", "0179427d4c6eb4f60db7e7ffbd57dfcd92378a38"},
	{"refs/tags/beta/v0.0.1.beta", "b8f7817589b8d09bf494100a16f80a6bc0e8a1f0"},
	{"refs/tags/beta/v0.0.2.beta", "c4f090f1eca10e57cc143d73adb987f08fa49184"},
}

var tagExamples = []refExample{
	{
		Example:  "this/is/a/path",
		Expected: []*refInfo{},
	},
	{"v0.0.2", []*refInfo{
		newRefLink("v0.0.2", cp(1, "0179427d4c6eb4f60db7e7ffbd57dfcd92378a38"))},
	},
	{"", []*refInfo{
		newRefLink("v0.0.1", cp(1, "5a57370fe617240f839a8a2e41f3b23c97f008fb")),
		newRefLink("v0.0.2", cp(1, "0179427d4c6eb4f60db7e7ffbd57dfcd92378a38")),
		newRefDir("beta")},
	},
	{"beta", []*refInfo{
		newRefLink("v0.0.1.beta", cp(2, "b8f7817589b8d09bf494100a16f80a6bc0e8a1f0")),
		newRefLink("v0.0.2.beta", cp(2, "c4f090f1eca10e57cc143d73adb987f08fa49184"))},
	},
}

func TestGetRefsTag(t *testing.T) {
	require := require.New(t)

	for _, e := range tagExamples {
		iter := newMockRefIter(mockTags)
		refs, err := getRefs(iter, e.Example)
		require.NoError(err)

		got := refExample{
			Example:  e.Example,
			Expected: refs,
		}

		require.Equal(e, got)
	}
}
