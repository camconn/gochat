package main

import (
	"container/list"
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
		Mode:  "",
		Topic: "",
		Users: list.New(),
	}

	return &c
}
