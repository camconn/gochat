package main

import (
	// "bytes"
	"container/list"
	"log"
	"net"
)

const bufSize = 1400
const CRLF = "\x0D\x0A"

type Event struct {
	Addr net.Addr
	Type int
	User *Client
	Body string
}

type Channel struct {
	Name  string
	Mode  string
	Topic string
	Users list.List
}

type Client struct {
	Conn      net.Conn
	Nick      string
	Username  string
	Type      int
	LastSeen  int64
	SendQueue list.List
	Realname  string
	Mode      string
}

type Message struct {
	Text string
	User *Client
}

type Packet struct {
	Target *Client
	Text   string
}

func main() {
	log.Println("Starting Server")
}
