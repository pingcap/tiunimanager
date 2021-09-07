package embed

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pingcap/check"
)

func Test(t *testing.T) { check.TestingT(t) }

type embedSuite struct{}

var _ = check.Suite(&embedSuite{})

func getAllFilePaths(dir string) (paths []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			return nil
		}
		if info.IsDir() {
			subPaths, err := getAllFilePaths(path)
			if err != nil {
				return err
			}
			paths = append(paths, subPaths...)
		} else {
			paths = append(paths, path)
		}

		return nil
	})

	return
}

// Test can read all file in /templates
func (s *embedSuite) TestCanReadConfigs(c *check.C) {
	paths, err := getAllFilePaths("configs")
	c.Assert(err, check.IsNil)
	c.Assert(len(paths), check.Greater, 0)

	for _, path := range paths {
		c.Log("check file: ", path)

		data, err := os.ReadFile(path)
		c.Assert(err, check.IsNil)

		embedData, err := ReadConfigTemplate(path)
		c.Assert(err, check.IsNil)

		c.Assert(embedData, check.BytesEquals, data)
	}
}

// Test can read all file in /examples
func (s *embedSuite) TestCanReadScripts(c *check.C) {
	paths, err := getAllFilePaths("scripts")
	c.Assert(err, check.IsNil)
	c.Assert(len(paths), check.Greater, 0)

	for _, path := range paths {
		c.Log("check file: ", path)

		data, err := os.ReadFile(path)
		c.Assert(err, check.IsNil)

		embedData, err := ReadScriptTemplate(path)
		c.Assert(err, check.IsNil)

		c.Assert(embedData, check.BytesEquals, data)
	}
}
