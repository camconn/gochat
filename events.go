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
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
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

	CONNECT
	HELP
	JOIN
	MODE
	MOTD
	MSG
	NICK
	PART
	PASS
	PING
	PONG
	QUIT
	REGISTERED
	RULES
	TOPIC
	USER
	VERSION_SERVER
)

const SPACE = " "
const COLON = ":"
const COMMA = ","

// Match Alphanumeric for first character, and Alphanumeric for the rest
// along with: .[]()-_
// Max nick length: 16 characters
const NICKREGEX = "^[A-Za-z0-9]([A-Za-z0-9\\.\\[\\]\\(\\)\\-\\_]){0,15}$"

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
	case "help":
		e.Type = HELP

		if len(words) >= 2 {
			e.Body = strings.Trim(raw[start:], SPACE+COLON)
		}
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

		if len(chanReasonPair) == 2 {
			e.Body = strings.Trim(chanReasonPair[1], COLON+SPACE) // leave reason
		}
	case "pass":
		e.Type = PASS
	case "ping":
		e.Type = PING
		pair := strings.Split(raw, SPACE)

		if len(pair) == 2 {
			e.Body = strings.Trim(pair[1], COLON+SPACE)
		}
	case "pong":
		e.Type = PONG
		words := strings.Split(raw, SPACE)

		l := len(words)
		if l >= 2 {
			lagStr := strings.Trim(words[l-1], COLON+SPACE)

			if len(lagStr) >= 16 {
				e.Body = "Y"
			}
		}
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

		pair := strings.SplitAfterN(raw[start:], COLON, 2)

		if len(pair) == 2 {
			e.Body = strings.Trim(pair[1], COLON+SPACE)
		}
	case "rules":
		e.Type = RULES
	case "topic":
		e.Type = TOPIC

		if len(words) == 3 || len(words) == 2 {
			e.Target = words[1]
			pair := strings.SplitAfterN(raw, COLON, 2)

			if len(pair) == 2 {
				e.Body = pair[1]
			}
			// else: blank topic
		} else {
			e.Valid = false
		}
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
	case "version":
		e.Type = VERSION_SERVER
	default:
		e.Type = UNKNOWN
	}

	return &e
}

func eventHandler(s *ServerInfo, events <-chan *Event) {
	channels := make(map[string]*Channel)
	users := make(map[string]*Client)

	nickRegex, _ := regexp.Compile(NICKREGEX)

	for {
		e := <-events
		fmt.Printf("Got event %v\n", e)

		switch e.Type {
		case HELP:
			log.Println("Help event")
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
						if i == -1 {
							continue
						} else {
							e.Sender.Channels = append(e.Sender.Channels, v)
							sort.Strings(e.Sender.Channels)
							log.Println("Adding user to channel", v)
						}
					} else {
						e.Sender.Channels = append(e.Sender.Channels, v)
					}

					if c, exists := channels[v]; exists {
						// add user to existing channel
						c.Users.PushBack(e.Sender)
					} else {
						// time to make a new channel
						channels[v] = NewChannel(v)
						channels[v].Users.PushBack(e.Sender)
					}

					channels[v].sendEvent(e.Sender, "JOIN", "")

					if len(channels[v].Topic) > 0 {
						e.Sender.sendServerTargetInfo(s, RPL_TOPIC, v, channels[v].Topic)
					} else {
						e.Sender.sendServerTargetInfo(s, RPL_NOTOPIC, v, "No topic is set")
					}
					channels[v].nameReply(s, e.Sender)
				}
			}
		case MODE:
			log.Println("Mode event")

			l := len(e.Target)

			if l == 0 {
				e.Sender.sendServerTargetInfo(s, ERR_NEEDMOREPARAMS, "MODE", "Need more parameters")
				continue
			}

			if l > 1 && (e.Target[0] == '#' || e.Target[0] == '&') { // sending to channel
				if ch, exists := channels[e.Target]; exists {
					// Format is:
					// :Server 324 nick #channel +modes
					e.Sender.sendMessage(strings.Join([]string{
						s.Hostname,
						padNumeric(RPL_CHANNELMODEIS),
						e.Sender.Nick,
						e.Target,
						"+" + ch.Mode,
					}, SPACE))
				} else {
					e.Sender.sendServerTargetInfo(s, ERR_NOSUCHCHANNEL, e.Target, "No such channel")
				}
			} else if l > 1 { // sending to user
				if user, exists := users[e.Target]; exists {
					// Print user mode info
					e.Sender.sendMessage(strings.Join([]string{
						s.Hostname,
						padNumeric(RPL_UMODEIS),
						e.Sender.Nick,
						e.Target,
						"+" + user.Mode,
					}, SPACE))
				} else {
					e.Sender.sendServerTargetInfo(s, ERR_NOSUCHNICK, e.Target, "No suck nick")
				}
			} else {
				log.Println("I shouldn't be here!")
			}

		case NICK:
			log.Println("User nick event")

			n := e.Body

			if !nickRegex.MatchString(n) {
				e.Sender.sendServerMessage(s, ERR_ERRONEUSNICKNAME, "Erroneus nickname.")
				continue
			}

			u, exists := users[n]

			if exists && u != e.Sender {
				log.Println("User already exists!")
				e.Sender.sendServerMessage(s, ERR_NICKNAMEINUSE, "Nickname is already in use.")
			} else if exists && u == e.Sender {
				// if user is changing their nick to what their nick already is, then ignore
				continue
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
					if binarySearch(ch, e.Sender.Channels) != -1 {
						partedChannel.sendEvent(e.Sender, "PART", reason)
					} else {
						e.Sender.sendServerMessage(s, ERR_NOTONCHANNEL, "You can't leave a channel you aren't in.")
					}
				} else {
					e.Sender.sendServerMessage(s, ERR_NOSUCHCHANNEL, "That channel does not exist")
				}
			}
		case PING:
			log.Println("Got PING, sending PONG")

			e.Sender.LastSeen = time.Now().Unix()

			if len(e.Body) > 0 {
				e.Sender.sendMessage(s.Hostname + " PONG " + s.Hostname + " :" + e.Body)
			} else {
				e.Sender.sendMessage(s.Hostname + " PONG " + s.Hostname + " :")
			}
		case PONG:
			e.Sender.LastSeen = time.Now().Unix()
			log.Println("Got PONG")
		case MSG:
			log.Println("Message event")

			l := len(e.Target)

			if l > 1 && (e.Target[0] == '#' || e.Target[0] == '&') { // sending to channel
				if c, exists := channels[e.Target]; exists {

					i := binarySearch(strings.ToLower(e.Target), e.Sender.Channels)

					restricted := c.hasMode(NO_EXTERNAL_MESSAGES)

					if i == -1 && restricted {
						e.Sender.sendServerTargetInfo(s, ERR_CANNOTSENDTOCHAN, e.Target,
							"Cannot send to channel (you need to join first)")
					} else {
						c.sendEvent(e.Sender, "PRIVMSG", e.Body)
					}
				} else {
					e.Sender.sendServerTargetInfo(s, ERR_CANNOTSENDTOCHAN, e.Target, "Cannot send to channel")
				}
			} else if l > 1 { // sending to user
				if user, exists := users[e.Target]; exists {
					user.sendMessage(e.Sender.Nick + " PRIVMSG " + user.Nick + " :" + e.Body)
				} else {
					e.Sender.sendServerTargetInfo(s, ERR_NOSUCHNICK, e.Target, "No suck nick")
				}
			} else { // invalid target
				e.Sender.sendServerMessage(s, ERR_NORECIPIENT, "No recipient given (PRIVMSG)")
			}
		case MOTD:
			log.Println("MOTD event")
			e.Sender.sendMotd(s)
		case RULES:
			log.Println("Rules event")
			// TODO: Actualy send rules
		case TOPIC:
			// TODO: Check if user has permission to change topic
			log.Println("TOPIC event")

			if !e.Valid {
				// TODO: Send bad command text
				continue
			}

			if ch, exists := channels[e.Target]; exists {
				ch.Topic = e.Body

				ch.sendEvent(e.Sender, "TOPIC", e.Body)
			} else {
				e.Sender.sendServerMessage(s, ERR_NOSUCHCHANNEL, e.Target+": No such channel")
			}
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

					e.Sender.Ping(s)
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

			// cleanup - disable all further messages and close connection.
			e.Sender.Alive = false
			e.Sender.Conn.Close()

			for _, ch := range channels {
				ch.sendEvent(e.Sender, "QUIT", e.Body)
				ch.removeUser(e.Sender.Nick)
			}

			// remove user from users map
			delete(users, e.Sender.Nick)
		case VERSION_SERVER:
			log.Println("Version event")
			e.Sender.sendVersion(s)
		case UNKNOWN:
		default:
			e.Sender.sendServerMessage(s, ERR_UNKNOWNCOMMAND, "Unknown command")
			log.Println("UNKNOWN event type.")
		}
	}
}
