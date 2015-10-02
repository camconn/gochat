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
	"log"
	"net"
	"strings"
)

type Client struct {
	Conn       net.Conn
	Cloak      string
	Channels   []string
	Nick       string
	Username   string
	Type       int
	LastSeen   int64 // TODO: Update on PINGs, PRIVMSG, JOIN, etc.
	Realname   string
	Mode       string
	Alive      bool // NOTE: Is this even needed? It is hardly ever used
	Registered bool
}

func (c *Client) sendMessage(message string) {
	c.sendRaw(":" + message)
}

// Send message to user and append CRLF to the end of the message.
// Checks if user's connection is active as a double-check
func (c *Client) sendRaw(message string) {
	go func(cl *Client) {
		if cl.Alive {
			log.Println(message)
			c.Conn.Write([]byte(message + CRLF))
		}
	}(c)
}

// Send a simple server numeric message in the format of
// :HOSTNAME 123 USERNICK :MESSAGE
func (c *Client) sendServerMessage(s *ServerInfo, numeric int, message string) {
	c.sendMessage(s.Hostname + " " + padNumeric(numeric) + " " + c.Nick + " :" + message)
}

// Send a user information (such as a topic, user list, or ERR_NOSUCHNICK error) about a
// target, which can be either a Channel, Nickname, or Server
func (c *Client) sendServerTargetInfo(s *ServerInfo, numeric int, target, message string) {
	c.sendMessage(s.Hostname + " " + padNumeric(numeric) + " " + c.Nick + " " + target + " :" + message)
}

func (c *Client) String() string {
	if len(c.Cloak) > 0 {
		return c.Nick + "!" + c.Username + "@" + c.Cloak
	} else {
		return c.NoCloakString()
	}
}

// Print out a user's nick, username, and host exposing personally-identifiable information
func (c *Client) NoCloakString() string {
	return c.Nick + "!" + c.Username + "@" + strings.Split(c.Conn.RemoteAddr().String(), COLON)[0]
}

func NewClient(connection net.Conn) Client {
	log.Println("New client: ", connection.RemoteAddr().String())
	c := Client{
		Conn:  connection,
		Cloak: "",
		Alive: true,
	}

	return c
}
