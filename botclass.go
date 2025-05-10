package main

import (
	"botTtrader/Items"
	"botTtrader/Users"
)

type Bot struct {
	Owner   *Users.User `json:"owner"`
	Channel string      `json:"channel"`
}

type Store struct {
	Name  string        `json:"name"`
	Items []*Items.Item `json:"items"`
	Owner *Users.User   `json:"owner"`
}
