# Memcached Server

## Overview

A custom memcached server written in Go.

### Build

Build the service:
```bash
$ go build -o ccmemcached cmd/memcached/main.go
```

### Run

The server runs on `127.0.0.1:11211` by default:

```bash
$ ./ccmemcached
```

You can specify a different address and port to use for running the server:

```bash
$ ./ccmemcached -a 192.168.0.1 -p 11212
```


## Run

```bash

```

## Acknowledgement

Coding Challenge ["Build Your Own Memcached Server"](https://codingchallenges.fyi/challenges/challenge-memcached/). Go check out John Crickett's [Coding Challenges newsletter](https://codingchallenges.fyi/) for more inspiring challenges.
