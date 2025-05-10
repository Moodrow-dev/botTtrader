package main

import (
	"botTtrader/Customer"
	"botTtrader/Items"
	"database/sql"
	"log"
	"time"
)

func main() {
	db, err := sql.Open("sqlite3", "./store.db")
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	Items.InitDB(db)
	Items.Save(Items.Item{1, "Pin", "Товар 1", "Первый пробный товар", 100, "123", time.Now()}, db)
	_, bh, _ := createBotAndPoll()

	Customer.Menu(bh)
	Customer.Price(bh, db)
	Customer.GetInfoAboutItem(bh, db)

	go func() {
		err1 := bh.Start()
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	select {}
}
