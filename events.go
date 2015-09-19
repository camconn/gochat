package main

import (
	"fmt"
	"log"
	"sort"
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
	MODE
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
const COMMA = ","

// Create a new Event from a sending client and the raw command string
// The sole purpose of this function is the create an Event object and
// specify the proper body, target, and do a simple preliminary check of
// if the Event message is valid.
func NewEvent(cl *Client, raw string) *Event {
	e := Event{
		Type:   UNKNOWN,
		Sender: cl,
		Target: "",
		Body:   "",
		Valid:  true,
	}

	// if raw == "", a blank event is made
	if raw == "" {
		return &e
	}

	log.Println(raw)

	words := strings.Split(raw, SPACE)
	start := len(words[0]) + 1 // index of first char of first word

	var command string

	if len(words) >= 2 {
		command = strings.ToLower(words[0])
	} else if len(words) == 1 {
		command = strings.ToLower(raw)
	}

	fmt.Printf("Words: %v\n", words)

	switch command {
	case "join":
		e.Type = JOIN
		e.Body = strings.Trim(raw[start:], SPACE)
	case "mode":
		e.Type = MODE
		e.Body = strings.Trim(raw[start:], SPACE)

		if len(words) >= 2 {
			e.Target = strings.ToLower(words[1])
		} else {
			e.Valid = false
		}

	case "motd":
		e.Type = MOTD
	case "nick":
		e.Type = NICK

		if len(words) == 2 {
			e.Body = strings.Trim(words[1], SPACE)
		} else {
			e.Valid = false
		}
	case "part":
		e.Type = PART

		chanReasonPair := strings.SplitAfterN(raw[start:], COLON, 2)

		e.Target = strings.Trim(chanReasonPair[0], COLON+SPACE) // comma-separated list of channels
		e.Body = strings.Trim(chanReasonPair[1], COLON+SPACE)   // leave reason
	case "pass":
		e.Type = PASS
	case "ping":
		e.Type = PING
	case "privmsg":
		e.Type = MSG
		fmt.Printf("Private message: %s\n", raw[start:])
		targetMessagePair := strings.SplitAfterN(raw[start:], COLON, 2)

		if len(targetMessagePair) != 2 {
			e.Valid = false
			break
		}

		e.Target = strings.Trim(targetMessagePair[0], COLON+SPACE)
		e.Body = strings.Trim(targetMessagePair[1], COLON+SPACE)
	case "quit":
		e.Type = QUIT
	case "user":
		e.Type = USER
		log.Println("User received")
		fmt.Printf("Raw: %s\n", raw[start:])

		if len(words) >= 5 { // Use >= because some REALNAMEs have spaces in them
			e.Body = raw[start:]
			fmt.Printf("e.Body: %s\n", e.Body)
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
	channels := make(map[string]*Channel)
	users := make(map[string]*Client)

	for {
		e := <-events
		fmt.Printf("Got event %v\n", e)

		switch e.Type {
		case JOIN:
			log.Println("Join event")
			chanPassPair := strings.Split(e.Body, SPACE)

			log.Println(chanPassPair)

			if len(chanPassPair) > 2 || len(chanPassPair) < 0 {
				// TODO: send error message
			} else {
				chans := strings.Split(chanPassPair[0], COMMA)
				// keys := strings.Split(chanPassPair[1], COMMA)

				for _, v := range chans {
					v := strings.Trim(v, SPACE)

					log.Println("in loop")

					if len(v) == 0 || (v[0] != '#' && v[0] != '&') {
						log.Println("Invalid channel name", v)
						e.Sender.sendServerMessage(s, ERR_NOSUCHCHANNEL, "The channel \""+v+"\" does not exist")
						continue
					}

					// check if user is already in channel
					if len(e.Sender.Channels) > 0 {
						i := binarySearch(v, e.Sender.Channels)

						// Do nothing, the user is already in this channel
						if i != -1 {
							continue
						} else {
							e.Sender.Channels = append(e.Sender.Channels, v)
							sort.Strings(e.Sender.Channels)
							log.Println("Adding user to channel", v)
						}
					}

					if c, exists := channels[v]; exists {
						// add user to existing channel
						c.Users.PushBack(e.Sender)
					} else {
						// time to make a new channel
						channels[v] = NewChannel(v)
						channels[v].Users.PushBack(e.Sender)
					}

					e.Sender.sendServerChannelInfo(s, RPL_TOPIC, v, channels[v].Topic)
					channels[v].sendEvent(e.Sender, "JOIN", "")
					channels[v].nameReply(s, e.Sender)
				}
			}
		case MODE:
			log.Println("Mode event")
			fmt.Printf("e.Body: %s\n", e.Body)

			// e.Sender.sendServerChannelInfo
		case NICK:
			// TODO: check if selected nick is valid
			log.Println("User nick event")

			n := e.Body

			u, exists := users[n]

			if exists {
				log.Println("User already exists!")
				e.Sender.sendServerMessage(s, ERR_NICKNAMEINUSE, "Nickname is already in use.")
			} else {
				if e.Sender.Nick != "" {
					log.Println("User changed their nickname to", n)
					delete(users, n)
					users[n] = u
				} else { // User is connecting for first time
					log.Println("New user: ", n)
					users[n] = e.Sender
				}

				e.Sender.Nick = n
			}
		case PART:
			log.Println("Leave channel event")

			chans := strings.Split(e.Target, COMMA) // e.Target is a comma-separated list of channels
			reason := strings.Trim(e.Body, SPACE)   // e.Body is the part reason

			for _, ch := range chans {
				if partedChannel, exists := channels[ch]; exists {
					partedChannel.sendEvent(e.Sender, "PART", reason)
				} else {
					e.Sender.sendServerMessage(s, ERR_NOSUCHCHANNEL, "That channel does not exist")
				}
			}
		case PING:
			log.Println("Got PING, sending PONG")
			e.Sender.sendRaw("PONG " + s.Hostname)
		case MSG:
			log.Println("Message event")

			// TODO: Check if PRIVMSG is being sent to a user
			if c, exists := channels[e.Target]; exists {
				// TODO: Check if user has joined channel
				c.sendEvent(e.Sender, "PRIVMSG", e.Body)
			} else {
				e.Sender.sendServerMessage(s, ERR_CANNOTSENDTOCHAN, "Cannot send to channel")
			}
		case MOTD:
			log.Println("MOTD event")
			e.Sender.sendMotd(s)
		case USER:
			log.Println("User info event")

			nick := e.Sender.Nick

			if len(nick) > 1 {
				parts := strings.SplitAfterN(e.Body, SPACE, 4)

				if len(parts) == 4 {
					e.Sender.Username = strings.Trim(parts[0], SPACE)
					e.Sender.Realname = strings.Trim(parts[3], SPACE+COLON)
					e.Sender.Registered = true

					log.Println("User information registered for", e.Sender.Realname)

					e.Sender.sendWelcomeMessage(s)
				} else {
					// TODO: Send INVALID error
					log.Println("Invalid USER command")
					log.Printf("%v\n", parts)
				}
			} else {
				log.Println("User must first send a username")
			}
		case QUIT:
			log.Println("User quit event from ", e.Sender.Nick)

			// close connections
			e.Sender.Conn.Close()

			// remove user from users map
			delete(users, e.Sender.Nick)

			// TODO: remove user from all participating channels
		case UNKNOWN:
		default:
			e.Sender.sendServerMessage(s, ERR_UNKNOWNCOMMAND, "Unknown command")
			log.Println("lol, don't know what type of event this is!")
		}
	}
}
