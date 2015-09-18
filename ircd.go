package main

import (
	"bytes"
	"container/list"
	"log"
	"net"
	"strings"
)

const bufSize = 1400
const CRLF = "\x0D\x0A"

type Client struct {
	Conn       net.Conn
	Nick       string
	Username   string
	Type       int
	LastSeen   int64
	SendQueue  list.List
	Realname   string
	Mode       string
	Alive      bool
	Registered bool
}

type Message struct {
	Text string
	User *Client
}

func (c *Client) sendMessage(message string) {
	go func() {
		c.Conn.Write([]byte(":" + message + CRLF))
	}()
}

func (c *Client) String() string {
	return c.Nick + "!" + c.Username + "@" + strings.Split(c.Conn.RemoteAddr().String(), COLON)[0]
}

func NewClient(connection net.Conn) Client {
	log.Println("New client: ", connection.RemoteAddr().String())
	c := Client{
		Conn:  connection,
		Alive: true,
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

	go eventHandler(events)
	// TODO: Spawn Processor thread

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Couldn't accept connection: ", err)
		}

		cl := NewClient(conn)
		go handleConnection(&cl, msgsIn, events)
	}
}

func handleConnection(cl *Client, in chan string, events chan<- *Event) {
	var bufferIn []byte
	msgBuffer := make([]byte, 0)
	log.Println("Now handling new connection")

	for cl.Alive {
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
			// Go ahead and convert to strings while we're at it
			l := len(msg)
			if l > 0 {
				msgString := string(msg[:l])
				events <- NewEvent(cl, msgString)
			}
		}

		msgBuffer = []byte{}
	}
}

func main() {
	log.Println("Starting Server")

	// TODO: Load configuration from file

	networkHandler()
}
