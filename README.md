wc-lang-packs-server
====================

`wc-lang-packs-server`k serves the translation API for WooCommerce extension and
language packs (zip file containing .mo and .po files).

## Quick Install

First, you need to:

* Install [Go](https://golang.org/doc/install)
* Clone this repo

Inside this repo:

```
go build
```

## Usage

See `-h` for available flags.

```
$ wc-lang-packs-server

Usage of wc-lang-packs-server:
  -db string
       	Full path to DB file (default "/var/folders/lk/xzr9s0655d3f54p8t2p2p45h0000gn/T/wc-lang-packs.db")
  -downloads-path string
       	Full path to serve language packs files (default "/var/folders/lk/xzr9s0655d3f54p8t2p2p45h0000gn/T/downloads")
  -exposedb
       	Expose /_db/ to dump in-memory DB as JSON
  -gpApiURL string
       	Root API project of GlotPress (default "https://translate.wordpress.com/api/projects/")
  -gpURL string
       	Root project of GlotPress (default "https://translate.wordpress.com/projects/")
  -listen string
       	HTTP listen address (default ":8081")
  -mode string
       	Check mode, 'poll' or 'notified' (default "poll")
  -poll-interval duration
       	Interval to poll translate.wordpress.com API if mode is poll (default 10m0s)
  -seed
       	Seed the DB before serving requests
  -update-key string
       	Key to post update if mode is notified (default "my-secret-key")
```
