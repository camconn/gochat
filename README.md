# gochat

A light and speedy IRC server written in Go.

### About

The goal of *gochat* is to be a lightweight (low memory usage) and fast (low
message latency) IRC server. To do this, *gochat* takes heavy advantage of
Go's goroutines and channels.

### Usage

To get a server up and running, use the following commands:
```
git clone https://github.com/camconn/gochat.git
cd gochat
go get
go build
./gochat
```

### Configuration

The configuration file for this program is found at `config.ini`. You can specify an 
alternative MOTD file by either changing the `motd` option, or by editing `motd.txt` yourself.

### License

This project is licensed under the GNU Public License, Version 3 or Later.
A copy of this license can be found in `LICENSE.md`.

### Disclosure

This project was inspired by the [goircd](https://github.com/stargrave/goircd) project,
which is very similar to what this project does. I copied the fashion in how messages 
are read from sockets (`handleConnection()` in `gochat.go`). This is a very small portion
of the code base, however, and should not effect the originality of *gochat*.
