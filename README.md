# IRCd

An IRC daemon written in Go.

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

The configuration file for this program is found at `config.ini`. You can specify an alternative MOTD file by either changing the `motd` option, or by editing `motd.txt` yourself.
