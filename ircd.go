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

func NewClient(connection net.Conn) Client {
	log.Println("New client: ", connection.RemoteAddr().String())
	c := Client{
		Conn: connection,
	}

	return c
}

func networkHandler() {
	listener, err := net.Listen("tcp", ":6667")

	if err != nil {
		log.Fatal("Couldn't listen on port 6667: ", err)
	}

	msgsIn := make(chan string)
	events := make(chan *Event)
	// packets := make(chan *Packet)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Couldn't accept connection: ", err)
		}

		cl := NewClient(conn)
		go handleConnection(&cl, msgsIn, events)
	}
}

func handleConnection(cl *Client, in chan string, events <-chan *Event) {
	return
}

func main() {
	log.Println("Starting Server")

	// TODO: Load configuration from file

	networkHandler()
}
