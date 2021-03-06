/*
gochat -- A light and speedy IRC server.
Copyright (C) 2015 Cameron Conn <cam_at_camconn_dot_cc>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"log"
	"net"
)

const bufSize = 1400
const CRLF = "\x0D\x0A"
const VERSION = "0.0.2-alpha"

// Accept connections and hand off each new client to a separate
// goroutine
func networkHandler(s *ServerInfo) {
	listener, err := net.Listen("tcp", ":6667")

	if err != nil {
		log.Fatal("Couldn't listen on port 6667: ", err)
	}

	msgsIn := make(chan string)
	events := make(chan *Event)

	go eventHandler(s, events)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Couldn't accept connection: ", err)
		}

		cl := NewClient(conn)

		// cloak user if there is a default cloak
		if len(s.DefaultCloak) > 0 {
			cl.Cloak = s.DefaultCloak
		}

		go handleConnection(&cl, msgsIn, events)
	}
}

func handleConnection(cl *Client, in chan string, events chan<- *Event) {
	var bufferIn []byte
	msgBuffer := make([]byte, 0)
	log.Println("Now handling connection :" + cl.String())

	for cl.Alive {
		bufferIn = make([]byte, bufSize)
		_, err := cl.Conn.Read(bufferIn)
		if err != nil {
			log.Println("User" + cl.String() + "disconnected")
			cl.Alive = false
			e := NewEvent(cl, "")
			e.Type = QUIT
			events <- e
			break
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

	conf := loadConfig()

	networkHandler(conf)
}
