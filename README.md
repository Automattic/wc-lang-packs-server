wc-lang-packs-server
====================

`wc-lang-packs-server` serves the translation API for WooCommerce extension and
language packs (zip file containing .mo and .po files).

## Quick Install

First, you need to:

* Install [Go](https://golang.org/doc/install)
* Clone this repo

Inside cloned repo:

```
go build -o server
./server -baseurl="http://localhost" -listen=":8081"
```

Your server now can be accessed:

```
curl -i http://localhost:8081/api/v1/plugins?slug=woocommerce-bookings&version=1.9.12

HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Date: Sat, 17 Sep 2016 05:03:59 GMT
Content-Length: 440

{"es_ES":{"language":"es_ES","last_modified":"","english_name":"Spanish (Spain)","native_name":"Español","package":"http://localhost/downloads/woocommerce-bookings/1.9.12/woocommerce-bookings-1.9.12-es_ES.zip"},"pt_BR":{"language":"pt_BR","last_modified":"","english_name":"Portuguese (Brazil)","native_name":"Português do Brasil","package":"http://localhost/downloads/woocommerce-bookings/1.9.12/woocommerce-bookings-1.9.12-pt_BR.zip"}}
```

## Usage

See `-h` for available flags.

```
$ wc-lang-packs-server -h

Usage of ./wc-lang-packs-server:
  -baseurl string
       	Base URL to access this server (default "https://translation.woocommerce.com")
  -db string
       	Full path to DB file (default "/tmp/wc-lang-packs/server.db")
  -downloads-path string
       	Full path to serve language packs files (default "/tmp/wc-lang-packs/downloads")
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
