# Memcached Server

## Overview

A custom memcached server written in Go.

## Build

You can build the service either directly, leveraging on `go build`, or with Docker:
```bash
$ go build -o ccmemcached cmd/main.go
```

```bash
$ docker build --no-cache=true -t ccmemcached .
```

## Run

You can run the server directly or leveaging on Docker.
By default, the server runs on `127.0.0.1:11211`:

```bash
$ ./ccmemcached
```

You can specify a different address and port to use for running the server:

```bash
$ ./ccmemcached -a 192.168.0.1 -p 11212
```

And with Docker
```bash
$ docker run -p 11211:11211 ccmemcached
```

To test its functioning, you can leverage on `telnet`:

```bash
$ telnet localhost 11211
Trying ::1...
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
set test1 0 0 4
1234
STORED
```

## Acknowledgement

Coding Challenge ["Build Your Own Memcached Server"](https://codingchallenges.fyi/challenges/challenge-memcached/). Go check out John Crickett's [Coding Challenges newsletter](https://codingchallenges.fyi/) for more inspiring challenges.
