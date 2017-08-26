# CheckupAPI

API server for Sourcegraph checkup data

[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/McKael/checkupapi)
[![license](https://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/McKael/checkupapi/master/LICENSE)
[![build](https://img.shields.io/travis/McKael/checkupapi.svg?style=flat)](https://travis-ci.org/McKael/checkupapi)
[![Go Report Card](https://goreportcard.com/badge/github.com/McKael/checkupapi)](https://goreportcard.com/report/github.com/McKael/checkupapi)

`checkupapi` is a [Go](https://golang.org/) REST API to access data generated by [checkup](https://github.com/sourcegraph/checkup).

## Purpose

The goal of checkupapi is to provide a common and efficient interface to read checkup's data without knowledge about checkup storage.  Currently, the supported checkup storages are those that implement the StorageReader interface (FS, SQL and Github).  (The Github backend has not been tested.)

With checkupapi, getting an event timeline or statistics is very fast and does not require a heavy CPU load on the client side.

Please check the [API specifications](CheckupAPI.md) for more details.

## Installation

You can install the API server with the go command line tool:

    go get github.com/McKael/checkupapi

and ugrade it with

    go get -u github.com/McKael/checkupapi

## Usage

checkupapi can use the same configuration file as checkup.

The default port is 8801, you can change it with the `--http` command line flag.

E.g.:

    checkupapi -c /etc/checkup/checkup.json --http :8080

## References

- [checkup](https://github.com/sourcegraph/checkup)
