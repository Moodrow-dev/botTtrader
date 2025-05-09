package main

type User struct {
	ID      int64  `bson:"id"`
	Name    string `bson:"name"`
	Phone   int    `bson:"phone"`
	Account string `bson:"account"`
	Address string `bson:"address"`
	isOwner bool
}

type Bot struct {
	Owner   User   `json:"owner"`
	Channel string `json:"channel"`
}

type Item struct {
	ID    int     `bson:"id"`
	Title string  `bson:"title"`
	Info  string  `bson:"info"`
	Price float64 `bson:"price"`
}

type Order struct {
	Customer *User  `bson:"customer"`
	Items    []Item `bson:"items"`
	ID       int    `bson:"id"`
	Track    string `bson:"track"`
	IsPaid   bool   `bson:"isPaid"`
}

type Store struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
	Owner *User  `json:"owner"`
}
