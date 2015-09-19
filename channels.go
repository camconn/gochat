package main

import (
	"container/list"
	"log"
)

type Channel struct {
	Name  string
	Mode  string
	Topic string
	Users *list.List
}

// Create a new chat channel
func NewChannel(name string) *Channel {
	c := Channel{
		Name:  name,
		Mode:  "+",
		Topic: "Default Topic",
		Users: list.New(),
	}

	log.Println("Creating new channel: " + name)

	return &c
}

// Send a message to all users in a channel
func (ch *Channel) sendToUsers(message string) {
	for user := ch.Users.Front(); user != nil; user = user.Next() {
		if c, ok := (user.Value).(*Client); ok {
			c.sendMessage(message)
		}
	}
}

// Send a user-generated action to all users in a channel with an optional message
// appended to the end
func (ch *Channel) sendEvent(sender *Client, action, message string) {
	for user := ch.Users.Front(); user != nil; user = user.Next() {
		if c, ok := (user.Value).(*Client); ok {
			if len(message) > 0 {
				c.sendRaw(":" + sender.String() + " " + action + " " + ch.Name + " :" + message)
			} else {
				c.sendRaw(":" + sender.String() + " " + action + " " + ch.Name)
			}
		}
	}
}

func (ch *Channel) nameReply(s *ServerInfo, recipient *Client) {
	users := []string{}
	for u := ch.Users.Front(); u != nil; u = u.Next() {
		if cl, ok := (u.Value).(*Client); ok {
			users = append(users, cl.Nick)
		} else {
			log.Println("ruh roh. `nil`  in channel user list")
		}
	}

	// cl.sendServerChannelInfo(s, RPL_NAMREPLY, ch.Name+" =", users)
}
