// Copyright (c) 2013 Jason McVetta.  This is Free Software, released under the
// terms of the AGPL v3.  See http://www.gnu.org/licenses/agpl-3.0.html for
// details. Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package main

import (
	"bytes"
	"crypto/rand"
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

// var _ = Suite(&TestSuite{})

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

func (s *TestSuite) TestReadWrite(c *C) {
	size := int(0.5 * KB)
	b := make([]byte, size)
	_, err := rand.Read(b)
	c.Assert(err, IsNil)
	resp, err := http.Post(s.url, "foobar", bytes.NewBuffer(b))
	defer resp.Body.Close()
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	log.Println(body)
}
