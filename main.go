// Copyright (c) 2013 Jason McVetta.  This is Free Software, released under the
// terms of the AGPL v3.  See http://www.gnu.org/licenses/agpl-3.0.html for
// details. Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/darkhelmet/env"
	"github.com/peterbourgon/diskv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const maxDataSize = int(64*MiB) / 8

// db is the key-value store to which data is persisted.
var db *diskv.Diskv

// Derived from http://golang.org/doc/effective_go.html#constants and
// https://groups.google.com/forum/#!topic/golang-nuts/AHoxOtHCOyw
type ByteSize float64

const (
	_            = iota
	KiB ByteSize = 1 << (10 * iota)
	MiB
	GiB
	TiB
	PiB
	EiB
	ZiB
	YiB
)

const transformBlockSize = 2 // grouping of chars per directory depth

func read(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get(":key")
	if key == "" {
		msg := "Must provide a key"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	b, err := db.Read(key)
	if os.IsNotExist(err) {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
	return
}

func write(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if len(b) > maxDataSize {
		msg := "Maximum block size is 64MiB"
		http.Error(w, msg, http.StatusRequestEntityTooLarge)
		return
	}
	h := sha1.New()
	h.Write(b)
	bs := h.Sum(nil)
	sum := base64.URLEncoding.EncodeToString(bs)
	if !db.Has(sum) {
		err = db.Write(sum, b)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	w.Write([]byte(sum))
}

func handler() http.Handler {
	m := pat.New()
	m.Get("/blocker/:key", http.HandlerFunc(read))
	m.Post("/blocker", http.HandlerFunc(write))
	return m
}

func setupDb(dbDir string) {
	// Based on https://github.com/peterbourgon/diskv/blob/master/examples/content-addressable-store/cas.go#L14
	blockTransform := func(s string) []string {
		sliceSize := len(s) / transformBlockSize
		pathSlice := make([]string, sliceSize)
		for i := 0; i < sliceSize; i++ {
			from, to := i*transformBlockSize, (i*transformBlockSize)+transformBlockSize
			pathSlice[i] = s[from:to]
		}
		return pathSlice
	}

	// Initialize a new diskv store, rooted at dbDir, with a 1MiB cache.
	db = diskv.New(diskv.Options{
		BasePath: dbDir,
		// Transform:    func(s string) []string { return []string{} },
		Transform:    blockTransform,
		CacheSizeMax: uint64(maxDataSize),
	})
	return
}

func main() {
	log.SetFlags(log.Lshortfile)

	// Gather configuration from environment
	port := env.StringDefault("PORT", "8080")
	pwd := env.String("PWD")
	dbDir := env.StringDefault("DB_DIR", pwd+"/db")

	// Start the service
	setupDb(dbDir)
	fmt.Printf("Starting server on localhost:%v\n", port)
	h := handler()
	log.Fatal(http.ListenAndServe(":"+port, h))
}
