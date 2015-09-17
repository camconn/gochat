package main

import (
	"net"
)

type Event struct {
	Addr net.Addr
	Type int
	User *Client
	Body string
}

func NewEvent(cl *Client, raw string) *Event {
	e := Event{
		User: cl,
	}

	return &e
}
