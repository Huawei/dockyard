package rados

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/containerops/dockyard/utils/setting"
)

// Test hooks up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type RadosDriverSuite struct {
	driver *radosdesc
}

var _ = Suite(&RadosDriverSuite{})

// Run once when the suite starts running
func (s *RadosDriverSuite) SetUpSuite(c *C) {
	if err := setting.SetConfig("../../conf/containerops.conf"); err != nil {
		c.Assert(err, IsNil)
	}
	radosDriver, err := s.driver.New()
	c.Assert(err, IsNil)
	s.driver = radosDriver.(*radosdesc)

	c.Log("test bucket started")
}

// Run once after all tests or benchmarks have finished running
func (s *RadosDriverSuite) TearDownSuite(c *C) {
	// Delete Objects
	objects, err := s.driver.listObjects()
	c.Assert(err, IsNil)

	for _, object := range objects {
		err = s.driver.Ioctx.Delete(object)
		c.Assert(err, IsNil)
	}

	c.Log("test bucket completed")
}

// Run before each test or benchmark starts running
func (s *RadosDriverSuite) SetUpTest(c *C) {
	// Ensures that storage drivers have no object
	objects, err := s.driver.listObjects()
	c.Assert(err, IsNil)

	for _, object := range objects {
		err = s.driver.Ioctx.Delete(object)
		c.Assert(err, IsNil)
	}
}

// Run once after all tests or benchmarks have finished running
func (s *RadosDriverSuite) TearDownTest(c *C) {
	err := os.RemoveAll("/tmp/test")
	c.Assert(err, IsNil)
}

// TestWriteRead1 tests a simple write-read workflow.
func (s *RadosDriverSuite) TestWriteRead1(c *C) {
	filename := "/tmp/test" + randomPath(32)
	contents := []byte("无言独上西楼，月如钩")
	s.writeReadCompare(c, filename, contents)
}

// TestWriteRead2 tests a simple write-read workflow with unicode data.
func (s *RadosDriverSuite) TestWriteRead2(c *C) {
	filename := "/tmp/test" + randomPath(32)
	contents := []byte("\xc3\x9f")
	s.writeReadCompare(c, filename, contents)
}

// TestWriteRead3 tests a simple write-read workflow with a small string.
func (s *RadosDriverSuite) TestWriteRead3(c *C) {
	filename := "/tmp/test" + randomPath(32)
	contents := randomContents(32)
	s.writeReadCompare(c, filename, contents)
}

// TestWriteRead4 tests a simple write-read workflow with 1MB of data.
func (s *RadosDriverSuite) TestWriteRead4(c *C) {
	filename := "/tmp/test" + randomPath(32)
	contents := randomContents(1024 * 1024)
	s.writeReadCompare(c, filename, contents)
}

// TestWriteReadNonUTF8 tests that non-utf8 data may be written to the storage
// driver safely.
func (s *RadosDriverSuite) TestWriteReadNonUTF8(c *C) {
	filename := "/tmp/test" + randomPath(32)
	contents := []byte{0x80, 0x80, 0x80, 0x80}
	s.writeReadCompare(c, filename, contents)
}

// TestTruncate tests that putting smaller contents than an original file does
// remove the excess contents.
func (s *RadosDriverSuite) TestTruncate(c *C) {
	filename := "/tmp/test" + randomPath(32)
	contents := randomContents(1024 * 1024)
	s.writeReadCompare(c, filename, contents)

	contents = randomContents(1024)
	s.writeReadCompare(c, filename, contents)
}

// TestReadNonexistent tests reading content from an empty path.
func (s *RadosDriverSuite) TestReadNonexistent(c *C) {
	filename := "/tmp/test" + randomPath(32)
	_, err := s.driver.Get(filename)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "rados"), Equals, true)
}

// TestDelete checks that the delete operation removes data from the storage
// driver
func (s *RadosDriverSuite) TestDelete(c *C) {
	filename := "/tmp/test" + randomPath(32)
	contents := randomContents(32)

	directory := path.Dir(filename)
	err := os.MkdirAll(directory, 0777)
	c.Assert(err, IsNil)
	err = ioutil.WriteFile(filename, contents, 0666)
	c.Assert(err, IsNil)

	_, err = s.driver.Save(filename)
	c.Assert(err, IsNil)

	err = s.driver.Delete(filename)
	c.Assert(err, IsNil)

	_, err = s.driver.Get(filename)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "rados"), Equals, true)
}

// TestDeleteNonexistent checks that removing a nonexistent key fails.
func (s *RadosDriverSuite) TestDeleteNonexistent(c *C) {
	filename := randomPath(32)
	err := s.driver.Delete(filename)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "rados"), Equals, true)
}

func (s *RadosDriverSuite) writeReadCompare(c *C, filename string, contents []byte) {
	directory := path.Dir(filename)
	err := os.MkdirAll(directory, 0777)
	c.Assert(err, IsNil)
	err = ioutil.WriteFile(filename, contents, 0666)
	c.Assert(err, IsNil)

	_, err = s.driver.Save(filename)
	c.Assert(err, IsNil)

	readContents, err := s.driver.Get(filename)
	c.Assert(err, IsNil)

	c.Assert(readContents, DeepEquals, contents)
}

var filenameChars = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
var separatorChars = []byte("._-")

func randomPath(length int64) string {
	path := "/"
	for int64(len(path)) < length {
		chunkLength := rand.Int63n(length-int64(len(path))) + 1
		chunk := randomFilename(chunkLength)
		path += chunk
		remaining := length - int64(len(path))
		if remaining == 1 {
			path += randomFilename(1)
		} else if remaining > 1 {
			path += "/"
		}
	}
	return path
}

func randomFilename(length int64) string {
	b := make([]byte, length)
	wasSeparator := true
	for i := range b {
		if !wasSeparator && i < len(b)-1 && rand.Intn(4) == 0 {
			b[i] = separatorChars[rand.Intn(len(separatorChars))]
			wasSeparator = true
		} else {
			b[i] = filenameChars[rand.Intn(len(filenameChars))]
			wasSeparator = false
		}
	}
	return string(b)
}

// randomBytes pre-allocates all of the memory sizes needed for the test. If
// anything panics while accessing randomBytes, just make this number bigger.
var randomBytes = make([]byte, 128<<20)

func init() {
	// increase the random bytes to the required maximum
	for i := range randomBytes {
		randomBytes[i] = byte(rand.Intn(2 << 8))
	}
}

func randomContents(length int64) []byte {
	return randomBytes[:length]
}
