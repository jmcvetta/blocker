// Copyright (c) 2013 Jason McVetta.  This is Free Software, released under the
// terms of the AGPL v3.  See http://www.gnu.org/licenses/agpl-3.0.html for
// details. Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	Suite(&TestSuite{})
	TestingT(t)
}

type TestSuite struct {
	dbDir string
	url   string
	hserv *httptest.Server
}

func (s *TestSuite) SetUpSuite(c *C) {
	log.SetFlags(log.Lshortfile)
}

func (s *TestSuite) SetUpTest(c *C) {
	var err error
	s.dbDir, err = ioutil.TempDir("", "blocker")
	c.Assert(err, IsNil)
	setupDb(s.dbDir)
	h := handler()
	s.hserv = httptest.NewServer(h)
	s.url = s.hserv.URL + "/blocker"
}

func (s *TestSuite) TearDownTest(c *C) {
	s.hserv.Close()
	os.RemoveAll(s.dbDir)
}

// write POSTs a random byte slice value to the API.
func (s *TestSuite) write(c *C) (key string, value []byte) {
	// Write
	size := int(2 * MiB)
	value = make([]byte, size)
	_, err := rand.Read(value)
	c.Assert(err, IsNil)
	resp, err := http.Post(s.url, "foobar", bytes.NewBuffer(value))
	defer resp.Body.Close()
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)
	b, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	key = string(b)
	return
}

// TestWriteRead tests a round-trip of writing data to the API, then retrieving
// the same data from its sha1.
func (s *TestSuite) TestWriteRead(c *C) {
	// Write
	key, sendValue := s.write(c)
	// Read
	url := s.url + "/" + string(key)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(err, IsNil)
	retValue, err := ioutil.ReadAll(resp.Body)
	c.Assert(retValue, DeepEquals, sendValue)
}

// TestAlreadyExists tests writing a block that already exists on disk.
func (s *TestSuite) TestAlreadyExists(c *C) {
	// Write
	_, sendValue := s.write(c)
	// And again...
	resp, err := http.Post(s.url, "foobar", bytes.NewBuffer(sendValue))
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)

}

// TestCorrupt tests server response when data is corrupted on disk.
func (s *TestSuite) TestCorrupt(c *C) {
	// Write
	key, _ := s.write(c)
	// Cause intentional corruption
	db.Write(string(key), []byte("foobar"))
	// Expect 500 error - desirable behavior?
	url := s.url + "/" + string(key)
	resp, err := http.Get(url)
	// defer resp.Body.Close()
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 500)
}

// TestNonexistent tests for 404 error when requesting a non existent sha1.
func (s *TestSuite) TestNonexistent(c *C) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	c.Assert(err, IsNil)
	key := base64.URLEncoding.EncodeToString(b)
	url := s.url + "/" + string(key)
	r, err := http.Get(url)
	c.Assert(err, IsNil)
	c.Assert(r.StatusCode, Equals, http.StatusNotFound)
}
