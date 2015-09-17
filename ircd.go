package main

import (
	"bytes"
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

func NewEvent(cl *Client, raw string) *Event {
	e := Event{
		User: cl,
	}

	return &e
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
	var bufferIn []byte
	msgBuffer := make([]byte, 0)

	for {
		bufferIn = make([]byte, bufSize)
		_, err := cl.Conn.Read(bufferIn)
		if err != nil {
			log.Println("Couldn't read client input: ", err)
		}

		// strip null chars
		bufferIn = bytes.TrimRight(bufferIn, "\x00")

		// append messages until the buffer ends in a newline
		msgBuffer = append(msgBuffer, bufferIn...)
		if !bytes.HasSuffix(msgBuffer, []byte(CRLF)) {
			continue
		}

		for _, msg := range bytes.Split(msgBuffer[:len(msgBuffer)-2], []byte(CRLF)) {
			if len(msg) > 0 {
				events <- NewEvent(cl, msg)
			}
		}
	}
}

func main() {
	log.Println("Starting Server")

	// TODO: Load configuration from file

	networkHandler()
}
