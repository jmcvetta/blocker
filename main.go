// Copyright (c) 2013 Jason McVetta.  This is Free Software, released under the
// terms of the AGPL v3.  See http://www.gnu.org/licenses/agpl-3.0.html for
// details. Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package main

import (
	"github.com/darkhelmet/env"
	"net/http"
	// "code.google.com/p/leveldb-go/leveldb"
	"fmt"
	"log"
)

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		readHandler(w, r)
		return
	case "POST":
		writeHandler(w, r)
		return
	}
	msg := "Only GET and POST are supported."
	http.Error(w, msg, 405)
}

func readHandler(w http.ResponseWriter, r *http.Request) {

}

func writeHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	port := env.StringDefault("PORT", "8080")
	http.HandleFunc("/blocker", handler)
	fmt.Printf("Starting server on localhost:%v\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
