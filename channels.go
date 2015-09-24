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
	"container/list"
	"log"
	"strings"
	"time"
)

const (
	NO_EXTERNAL_MESSAGES = "n"
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
		Mode:    "n",
		Topic:   "Default Topic",
		Users:   list.New(),
		Created: time.Now().Unix(),
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
		if c, ok := (user.Value).(*Client); ok && !(c == sender && action == "PRIVMSG") {
			if len(message) > 0 {
				c.sendRaw(":" + sender.String() + " " + action + " " + ch.Name + " :" + message)
			} else {
				c.sendRaw(":" + sender.String() + " " + action + " " + ch.Name)
			}
		}
	}
}

// Send list of users in channel to recipient. This uses the
// RPL_NAMREPLY numeric code.
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
			recipient.sendServerTargetInfo(s, RPL_NAMREPLY, "= "+ch.Name, strings.Join(users[i:], SPACE))
		} else {
			recipient.sendServerTargetInfo(s, RPL_NAMREPLY, "= "+ch.Name, strings.Join(users[i:end], SPACE))
		}
	}

	recipient.sendServerTargetInfo(s, RPL_ENDOFNAMES, ch.Name, "End of NAMES list")
}

func (ch *Channel) hasMode(mode string) bool {
	return strings.Index(ch.Mode, mode) != -1
}
