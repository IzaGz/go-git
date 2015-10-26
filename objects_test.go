package git

import (
	"encoding/base64"
	"io/ioutil"
	"time"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v2/internal"
)

type ObjectsSuite struct{}

var _ = Suite(&ObjectsSuite{})

var CommitFixture = "dHJlZSBjMmQzMGZhOGVmMjg4NjE4ZjY1ZjZlZWQ2ZTE2OGUwZDUxNDg4NmY0CnBhcmVudCBiMDI5NTE3ZjYzMDBjMmRhMGY0YjY1MWI4NjQyNTA2Y2Q2YWFmNDVkCnBhcmVudCBiOGU0NzFmNThiY2JjYTYzYjA3YmRhMjBlNDI4MTkwNDA5YzJkYjQ3CmF1dGhvciBNw6F4aW1vIEN1YWRyb3MgPG1jdWFkcm9zQGdtYWlsLmNvbT4gMTQyNzgwMjQzNCArMDIwMApjb21taXR0ZXIgTcOheGltbyBDdWFkcm9zIDxtY3VhZHJvc0BnbWFpbC5jb20+IDE0Mjc4MDI0MzQgKzAyMDAKCk1lcmdlIHB1bGwgcmVxdWVzdCAjMSBmcm9tIGRyaXBvbGxlcy9mZWF0dXJlCgpDcmVhdGluZyBjaGFuZ2Vsb2c="

func (s *ObjectsSuite) TestNewCommit(c *C) {
	data, _ := base64.StdEncoding.DecodeString(CommitFixture)

	o := &internal.RAWObject{}
	o.SetType(internal.CommitObject)
	o.Writer().Write(data)

	commit := &Commit{}
	c.Assert(commit.Decode(o), IsNil)

	c.Assert(commit.Hash.String(), Equals, "a5b8b09e2f8fcb0bb99d3ccb0958157b40890d69")
	c.Assert(commit.Tree.String(), Equals, "c2d30fa8ef288618f65f6eed6e168e0d514886f4")
	c.Assert(commit.Parents, HasLen, 2)
	c.Assert(commit.Parents[0].String(), Equals, "b029517f6300c2da0f4b651b8642506cd6aaf45d")
	c.Assert(commit.Parents[1].String(), Equals, "b8e471f58bcbca63b07bda20e428190409c2db47")
	c.Assert(commit.Author.Email, Equals, "mcuadros@gmail.com")
	c.Assert(commit.Author.Name, Equals, "Máximo Cuadros")
	c.Assert(commit.Author.When.Unix(), Equals, int64(1427802434))
	c.Assert(commit.Committer.Email, Equals, "mcuadros@gmail.com")
	c.Assert(commit.Message, Equals, "Merge pull request #1 from dripolles/feature\n\nCreating changelog\n")
}

var TreeFixture = "MTAwNjQ0IC5naXRpZ25vcmUAMoWKrTw4PtH/Cg+b3yMdVKAMnogxMDA2NDQgQ0hBTkdFTE9HANP/U+BWSp+H2OhLbijlBg5RcAiqMTAwNjQ0IExJQ0VOU0UAwZK9aiTqGrAdeGhuQXyL3Hw9GX8xMDA2NDQgYmluYXJ5LmpwZwDVwPSrgRiXyt8DrsNYrmDSH5HFDTQwMDAwIGdvAKOXcadlH5f69ccuCCJNhX/DUTPbNDAwMDAganNvbgBah35qkGonQ61uRdmcF5NkKq+O2jQwMDAwIHBocABYavVn0Ltedx5JvdlDT14Pt20l+jQwMDAwIHZlbmRvcgDPSqOziXT7fYHzZ8CDD3141lq4aw=="

func (s *ObjectsSuite) TestParseTree(c *C) {
	data, _ := base64.StdEncoding.DecodeString(TreeFixture)

	o := &internal.RAWObject{}
	o.SetType(internal.TreeObject)
	o.SetSize(int64(len(data)))
	o.Writer().Write(data)

	tree := &Tree{}
	c.Assert(tree.Decode(o), IsNil)

	c.Assert(tree.Entries, HasLen, 8)
	c.Assert(tree.Entries[0].Name, Equals, ".gitignore")
	c.Assert(tree.Entries[0].Mode.String(), Equals, "-rw-r--r--")
	c.Assert(tree.Entries[0].Hash.String(), Equals, "32858aad3c383ed1ff0a0f9bdf231d54a00c9e88")
}

func (s *ObjectsSuite) TestBlobHash(c *C) {
	o := &internal.RAWObject{}
	o.SetType(internal.BlobObject)
	o.SetSize(3)
	o.Writer().Write([]byte{'F', 'O', 'O'})

	blob := &Blob{}
	c.Assert(blob.Decode(o), IsNil)

	c.Assert(blob.Size, Equals, int64(3))
	c.Assert(blob.Hash.String(), Equals, "d96c7efbfec2814ae0301ad054dc8d9fc416c9b5")

	data, err := ioutil.ReadAll(blob.Reader())
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "FOO")
}

func (s *ObjectsSuite) TestParseSignature(c *C) {
	cases := map[string]Signature{
		`Foo Bar <foo@bar.com> 1257894000 +0100`: {
			Name:  "Foo Bar",
			Email: "foo@bar.com",
			When:  time.Unix(1257894000, 0),
		},
		`Foo Bar <> 1257894000 +0100`: {
			Name:  "Foo Bar",
			Email: "",
			When:  time.Unix(1257894000, 0),
		},
		` <> 1257894000`: {
			Name:  "",
			Email: "",
			When:  time.Unix(1257894000, 0),
		},
		`Foo Bar <foo@bar.com>`: {
			Name:  "Foo Bar",
			Email: "foo@bar.com",
			When:  time.Time{},
		},
		``: {
			Name:  "",
			Email: "",
			When:  time.Time{},
		},
		`<`: {
			Name:  "",
			Email: "",
			When:  time.Time{},
		},
	}

	for raw, exp := range cases {
		got := ParseSignature([]byte(raw))
		c.Assert(got.Name, Equals, exp.Name)
		c.Assert(got.Email, Equals, exp.Email)
		c.Assert(got.When.Unix(), Equals, exp.When.Unix())
	}
}
