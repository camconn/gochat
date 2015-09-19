package main

import (
	"container/list"
	"log"
	"strings"
)

type Channel struct {
	Name    string
	Mode    string
	Topic   string
	Users   *list.List
	Created int64
}

// Create a new chat channel
func NewChannel(name string) *Channel {
	c := Channel{
		Name:    name,
		Mode:    "+",
		Topic:   "Default Topic",
		Users:   list.New(),
		Created: 0,
	}

	log.Println("Creating new channel: " + name)

	return &c
}

// Remove user from users list. Currently this is a O(n) operation
// TODO: Research more efficient methods of handling users.
// TODO: Research using binary tree for user management
func (ch *Channel) removeUser(nick string) {
	nick = strings.ToLower(nick)
	for e := ch.Users.Front(); e != nil; e = e.Next() {
		if cl, ok := (e.Value).(*Client); ok {
			if strings.ToLower(cl.Nick) == nick { // found our user
				ch.Users.Remove(e)
				log.Println("Removed user nick from ", ch.Name)
				return
			}
		}
	}
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
		if c, ok := (user.Value).(*Client); ok && (c != sender || action == "JOIN") {
			if len(message) > 0 {
				c.sendRaw(":" + sender.String() + " " + action + " " + ch.Name + " :" + message)
			} else {
				c.sendRaw(":" + sender.String() + " " + action + " " + ch.Name)
			}
		}
	}
}

// Send list of users in channel to recipient
func (ch *Channel) nameReply(s *ServerInfo, recipient *Client) {
	users := []string{}
	for u := ch.Users.Front(); u != nil; u = u.Next() {
		if cl, ok := (u.Value).(*Client); ok {
			users = append(users, cl.Nick)
		} else {
			log.Println("ruh roh. `nil`  in channel user list")
		}
	}

	end := 0
	for i := 0; end < len(users); i += 8 {
		end += 8
		if end > len(users) {
			recipient.sendServerChannelInfo(s, RPL_NAMREPLY, "= "+ch.Name, strings.Join(users[i:], SPACE))
		} else {
			recipient.sendServerChannelInfo(s, RPL_NAMREPLY, "= "+ch.Name, strings.Join(users[i:end], SPACE))
		}
	}

	recipient.sendServerChannelInfo(s, RPL_ENDOFNAMES, ch.Name, "End of NAMES list")
}
