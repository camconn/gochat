package main

import (
	"fmt"
	"log"
	"strings"
)

type Event struct {
	Type   int
	Sender *Client
	Target string
	Body   string
	Valid  bool
}

const (
	UNKNOWN = iota
	JOIN
	NICK
	MOTD
	QUIT
	PART
	PING
	PASS
	MSG
	USER
	CONNECT
	REGISTERED
)

const SPACE = " "
const COLON = ":"

// Create a new Event from a sending client and the raw command string
func NewEvent(cl *Client, raw string) *Event {
	e := Event{
		Type:   UNKNOWN,
		Sender: cl,
		Target: "",
		Body:   "",
		Valid:  true,
	}

	log.Println(raw)

	// comPair[0] is everything before the text
	// comPair[1] is the text
	comPair := strings.SplitAfterN(raw, ":", 1)

	comLower := strings.ToLower(comPair[0])
	commandBlocks := strings.SplitAfterN(comLower, " ", 2)
	command := strings.Trim(commandBlocks[0], SPACE)
	fmt.Printf("Stuff \"%s\"\n", command)

	switch command {
	case "join":
		e.Type = JOIN
	case "motd":
		e.Type = MOTD
	case "nick":
		e.Type = NICK

		if len(commandBlocks) > 1 {
			e.Body = strings.Trim(commandBlocks[1], SPACE)
		} else {
			e.Valid = false
			log.Println("Invalid nick command")
		}
	case "part":
		e.Type = PART
	case "pass":
		e.Type = PASS
	case "ping":
		e.Type = PING
	case "privmsg":
		e.Type = MSG
	case "quit":
		e.Type = QUIT
	case "user":
		e.Type = USER

		if len(commandBlocks) == 2 {
			e.Body = commandBlocks[1]
		} else {
			e.Valid = false
			// TODO: Send invalid USER param code
		}
	default:
		e.Type = UNKNOWN
	}

	return &e
}

func eventHandler(s *ServerInfo, events <-chan *Event) {
	// channels := make(map[string]*Channel)
	users := make(map[string]*Client)

	for {
		e := <-events
		fmt.Printf("Got event %v\n", e)

		switch e.Type {
		case JOIN:
			log.Println("Join event")
		case NICK:
			log.Println("User nick event")

			n := e.Body

			_, exists := users[n]

			if exists {
				log.Println("User already exists!")
			} else {
				if e.Sender.Nick != "" {
					log.Println("User changed their nickname to", n)
				} else { // User is connecting for first time
					log.Println("New user: ", n)
					users[n] = e.Sender
				}

				e.Sender.Nick = n
			}
		case PART:
			log.Println("Leave channel event")
		case PING:
			log.Println("Got PING, sending PONG")
			e.Sender.sendRaw("PONG " + s.Hostname)
		case MSG:
			log.Println("Message event")
		case MOTD:
			log.Println("MOTD event")
			e.Sender.sendMotd(s)
		case USER:
			log.Println("User info event")

			uname := e.Sender.Nick

			if len(uname) > 1 {
				parts := strings.Split(e.Body, SPACE)

				if len(parts) == 4 {
					e.Sender.Username = strings.Trim(parts[1], SPACE)
					e.Sender.Realname = strings.Trim(parts[3], SPACE+COLON)
					e.Sender.Registered = true

					log.Println("User information registered for", e.Sender.Realname)

					e.Sender.sendWelcomeMessage(s)
				} else {
					// TODO: Send INVALID error
				}
			} else {
				log.Println("User must first send a username")
			}
		case QUIT:
			log.Println("User quit event from ", e.Sender.Nick)
			delete(users, e.Sender.Nick)
		case UNKNOWN:
		default:
			log.Println("lol, don't know what type of event this is!")
		}
	}
}
