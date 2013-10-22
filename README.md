blocker
=======

Blocker is an HTTP (but not particularly RESTful) API to write arbitrary blocks
of data to storage, and to retrieve them by SHA1 hash key.


## Status

EXPERIMENTAL - Not recommended for general use.

[![Build Status](https://travis-ci.org/jmcvetta/blocker.png?branch=master)](https://travis-ci.org/jmcvetta/blocker)
[![Build Status](https://drone.io/github.com/jmcvetta/blocker/status.png)](https://drone.io/github.com/jmcvetta/blocker/latest)
[![Coverage Status](https://coveralls.io/repos/jmcvetta/blocker/badge.png?branch=master)](https://coveralls.io/r/jmcvetta/blocker?branch=master)


## Installation

Blocker requires [Go](http://golang.org) 1 or higher.  Install with `go get`:

```
go get github.com/jmcvetta/blocker
```


## Usage

```bash
# Start blocker in the background
$ blocker &
[1] 17059
Starting server on localhost:8080

# POST data and get SHA1 digest as key
$ wget -qO - --post-data="foo bar baz" http://localhost:8080/blocker ; echo
x1Z-izniQo44v5ySJqxo3kxn3Dk=

# Retreive data using key
$ wget -qO - http://localhost:8080/blocker/x1Z-izniQo44v5ySJqxo3kxn3Dk= ; echo
foo bar baz
```


## Name

The name "Blocker" is a silly pun on the popular application
"[Docker](https://github.com/dotcloud/docker)" - to which Blocker has no
particular relationship or similarity - and "blocks of data".


## License

This is Free Software, released under the terms of the [AGPL
v3](http://www.gnu.org/licenses/agpl-3.0.html).
