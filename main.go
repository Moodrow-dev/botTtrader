package main

import (
	"botTtrader/Customer"
	"botTtrader/Items"
	"botTtrader/Orders"
	"botTtrader/Users"
	"botTtrader/Utils"
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

	err = Items.InitDB(db)
	if err != nil {
		log.Fatal(err)
	}

	err = Users.InitDB(db)
	if err != nil {
		log.Fatal(err)
	}

	err = Orders.InitDB(db)
	if err != nil {
		log.Fatal(err)
	}

	err = Items.Save(Items.Item{1, "Pin", "Товар 1", -1, "Первый пробный товар", 100, "123", time.Now()}, db)
	if err != nil {
		log.Fatal(err)
	}

	_, bh, _ := Utils.CreateBotAndPoll()

	Utils.DeleteThis(bh)

	Customer.Menu(bh, db)
	Customer.Catalog(bh, db)
	Customer.ItemInfo(bh, db)
	Customer.MyCart(bh, db)
	Customer.Cabinet(bh, db)
	Customer.ClearCart(bh, db)
	Customer.AddItemToCart(bh, db)

	go func() {
		err1 := bh.Start()
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	select {}
}
