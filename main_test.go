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
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
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

// write POSTs random data to the API.
func (s *TestSuite) write(c *C) (key string, value []byte, statusCode int) {
	bi, err := rand.Int(rand.Reader, big.NewInt(int64(maxDataSize-1)))
	c.Assert(err, IsNil)
	size := int(bi.Int64()) + 1
	// size := int(mbs * int(MiB))
	// size := int(2 * MiB)
	value = make([]byte, size)
	_, err = rand.Read(value)
	c.Assert(err, IsNil)
	// We're not doing anything server-side with the bodyType
	resp, err := http.Post(s.url, "", bytes.NewBuffer(value))
	defer resp.Body.Close()
	c.Assert(err, IsNil)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		c.Fatal("Expected status 200 or 201, but got", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	key = string(b)
	statusCode = resp.StatusCode
	return
}

// TestWriteRead tests a round-trip of writing data to the API, then retrieving
// the same data from its sha1.
func (s *TestSuite) TestWriteRead(c *C) {
	// Write
	key, sendValue, _ := s.write(c)
	// Read
	url := s.url + "/" + string(key)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(err, IsNil)
	retValue, err := ioutil.ReadAll(resp.Body)
	c.Assert(retValue, DeepEquals, sendValue)
}

// TestConcurrentWrites tests between 100 and 200 concurrent writes to the API.
func (s *TestSuite) TestConcurrentWrites(c *C) {
	count := runtime.NumCPU() * 25
	wg := sync.WaitGroup{}
	wg.Add(count)
	writer := func() {
		s.write(c)
		wg.Done()
	}
	for i := 0; i < count; i++ {
		go writer()
	}
	wg.Wait()
}

// TestConcurrentSameData tests between 100 and 200 concurrent writes of
// identical data to the API.
func (s *TestSuite) TestConcurrentSameData(c *C) {
	b, err := rand.Int(rand.Reader, big.NewInt(int64(100)))
	c.Assert(err, IsNil)
	count := 100 + int(b.Int64())
	wg := sync.WaitGroup{}
	wg.Add(count)
	bi, err := rand.Int(rand.Reader, big.NewInt(int64(int(MiB)-1)))
	c.Assert(err, IsNil)
	size := int(bi.Int64()) + 1
	value := make([]byte, size)
	_, err = rand.Read(value)
	c.Assert(err, IsNil)
	writer := func() {
		resp, err := http.Post(s.url, "", bytes.NewBuffer(value))
		defer resp.Body.Close()
		c.Assert(err, IsNil)
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			c.Fatal("Expected status 200 or 201, but got", resp.StatusCode)
		}
		// Done
		wg.Done()
	}
	for i := 0; i < count; i++ {
		go writer()
	}
	wg.Wait()
}

// TestGetNoKey tests for 400 error when trying to GET without a sha1 key.
func (s *TestSuite) TestGetNoKey(c *C) {
	r, err := http.Get(s.url)
	c.Assert(err, IsNil)
	// StatusMethodNotAllowed is apparently set by pat
	c.Assert(r.StatusCode, Equals, http.StatusMethodNotAllowed)
}

// TestAlreadyExists tests writing a block that already exists on disk.
func (s *TestSuite) TestAlreadyExists(c *C) {
	// Write
	_, sendValue, _ := s.write(c)
	// And again...
	resp, err := http.Post(s.url, "", bytes.NewBuffer(sendValue))
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)

}

// TestDiskFull tests for proper error handling when the disk to which the
// server writes data is full.
func (s *TestSuite) TestDiskFull(c *C) {
	log.Println("WARNING: TestDiskFull not yet implemented!")
}

// TestCorrupt tests server response when data is corrupted on disk.
func (s *TestSuite) TestCorrupt(c *C) {
	// Write
	key, _, _ := s.write(c)
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
