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
	RPL_MOTDSTART        = 375
	RPL_MOTD             = 372
	RPL_ENDOFMOTD        = 376
	ERR_NOSUCHNICK       = 401
	ERR_NOSUCHCHANNEL    = 403
	ERR_CANNOTSENDTOCHAN = 404
	ERR_TOOMANYCHANNELS  = 405
	ERR_WASNOSUCHNICK    = 406
	ERR_TOOMANYTARGETS   = 407
)

type ServerInfo struct {
	Hostname string
	Network  string
	MotdData []string
	started  *time.Time
}

// Reader server info from file and load it
func loadConfig() *ServerInfo {
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

	motdPath := serverSec.Key("motd").String()
	if len(motdPath) == 0 {
		log.Fatal("Invalid \"motd\" key in [server]")
	}

	readMotd(serverConfig, motdPath)

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

	log.Println("MOTD is " + strconv.Itoa(len(data)) + " bytes long")

	motdRaw := string(data[:len(data)])
	data = []byte{} // go ahead and clean out byte array

	s.MotdData = strings.Split(motdRaw, NEWLINE)
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
