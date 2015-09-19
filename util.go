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
	"github.com/go-ini/ini"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const NEWLINE = "\n"
const TIMEFORMAT = "Mon, Jan _2 2006 at 15:04:05 (MST)"
const MOTD_LENGTH = 3000

const (
	RPL_WELCOME          = 001
	RPL_YOURHOST         = 002
	RPL_CREATED          = 003
	RPL_MYINFO           = 004
	RPL_CHANNELMODEIS    = 324
	RPL_TOPIC            = 332
	RPL_NAMREPLY         = 353
	RPL_ENDOFNAMES       = 366
	RPL_MOTDSTART        = 375
	RPL_MOTD             = 372
	RPL_ENDOFMOTD        = 376
	ERR_NOSUCHNICK       = 401
	ERR_NOSUCHCHANNEL    = 403
	ERR_CANNOTSENDTOCHAN = 404
	ERR_TOOMANYCHANNELS  = 405
	ERR_WASNOSUCHNICK    = 406
	ERR_TOOMANYTARGETS   = 407
	ERR_UNKNOWNCOMMAND   = 421
	ERR_NICKNAMEINUSE    = 433
	ERR_NOTONCHANNEL     = 442
)

type ServerInfo struct {
	Hostname     string
	Network      string
	MotdData     []string
	DefaultCloak string
	started      *time.Time
}

// Search for a term in a sorted space. Returns the index
// of the string if found. Returns -1 if not found
// NB: To use this function, the array must be sorted
func binarySearch(term string, space []string) int {
	start := 0
	end := len(space) - 1
	mid := (start + end) / 2

	for {
		// fmt.Printf("New indexes ==> start: %d || mid: %d || end: %d\n", start, mid, end)
		result := compareString(term, space[mid])

		if result == 0 {
			return mid
		} else if result < 0 {
			end = mid
		} else if result > 0 {
			if start == mid { // edge case for examining last item
				start += 1
			} else {
				start = mid
			}
		} else {
			// wut happened here?
		}

		if start >= mid && mid >= end {
			return -1
		}

		mid = start + (end-start)/2
	}
	return -1
}

// Because go 1.4 doesn't have the feature in the `strings` library, an equivalent
// is defined here.
func compareString(x, y string) int {
	if x == y {
		return 0
	} else if x < y {
		return -1
	} else {
		return 1
	}
}

// Reader server info from file and load it
func loadConfig() *ServerInfo {
	log.Println("Loading configuration from `config.ini`")
	serverConfig := &ServerInfo{}
	now := time.Now()
	serverConfig.started = &now

	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatal("Couldn't read configuration file: ", err, "| exiting now...")
	}

	serverSec, err := cfg.GetSection("server")
	if err != nil {
		log.Fatal("Malformed configuration file: no \"server\" section")
	}

	serverConfig.Hostname = serverSec.Key("hostname").String()
	if len(serverConfig.Hostname) == 0 {
		log.Fatal("Invalid \"hostname\" key in [server]")
	}

	serverConfig.Network = serverSec.Key("network").String()
	if len(serverConfig.Network) == 0 {
		log.Fatal("Invalid \"network\" key in [server]")
	}

	serverConfig.DefaultCloak = strings.Trim(serverSec.Key("defaultcloak").String(), SPACE)

	motdPath := serverSec.Key("motd").String()
	if len(motdPath) == 0 {
		log.Fatal("Invalid \"motd\" key in [server]")
	}

	readMotd(serverConfig, motdPath)

	log.Println("Configuration Loaded")
	return serverConfig
}

// read motd from file and write data to ServerInfo
func readMotd(s *ServerInfo, path string) {
	f, err := os.Open(path)

	if err != nil {
		log.Fatal("Invalid motd path: " + path)
	}

	data := make([]byte, MOTD_LENGTH)

	_, err = f.Read(data)

	motdRaw := string(data[:len(data)])
	data = []byte{} // go ahead and clean out byte array

	s.MotdData = strings.Split(motdRaw, NEWLINE)

	log.Println("MOTD Loaded")
}

func (c *Client) sendWelcomeMessage(s *ServerInfo) {
	c.sendServerMessage(s, RPL_WELCOME, "Welcome to the Internet Relay Network "+c.String())
	c.sendServerMessage(s, RPL_YOURHOST, "Your host is "+s.Hostname+", running version "+string(VERSION))

	dateCreatedStr := s.started.Format(TIMEFORMAT)
	c.sendServerMessage(s, RPL_CREATED, "This server was stared on "+dateCreatedStr)

	infoStr := s.Hostname + strings.Join([]string{s.Hostname, VERSION, "+", "+"}, SPACE)
	c.sendServerMessage(s, RPL_MYINFO, infoStr)

	c.sendMotd(s)
}

func (c *Client) sendMotd(s *ServerInfo) {
	c.sendServerMessage(s, RPL_MOTDSTART, "- "+s.Hostname+" message of the day")
	for _, line := range s.MotdData {
		c.sendServerMessage(s, RPL_MOTD, line)
	}
	c.sendServerMessage(s, RPL_ENDOFMOTD, "End of MOTD command")

}

// pad a numeric to a length of 3 characters
// TODO: Make safe and error resistant for numerics accidentally over 999
func padNumeric(n int) string {
	currLen := len(strconv.Itoa(n))
	repeats := 3 - currLen

	return strings.Repeat("0", repeats) + strconv.Itoa(n)
}
