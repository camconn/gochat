# IRCd

An IRC daemon written in Go.

### About

The goal of IRCd is to be a fast (speedy messaging) and lightweight (low memory usage) IRC server.

### Usage

To get a server up and running, use the following commands:
```
git clone https://github.com/camconn/ircd.git
cd ircd
go get
go build
./ircd
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
are read from sockets (`handleConnection()` in `ircd.go`).

While this portion of the program *is* adopted from `goircd`, the two projects are sufficently
different to consider `ircd` a different work.
