package main

import (
	"botTtrader/Customer"
	"botTtrader/Utils"
	"database/sql"
	"log"
)

func main() {
	db, err := sql.Open("sqlite3", "./store.db")
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	//Items.InitDB(db)
	//Users.InitDB(db)
	//Orders.InitDB(db)
	//Items.Save(Items.Item{1, "Pin", "Товар 1", "Первый пробный товар", 100, "123", time.Now()}, db)
	_, bh, _ := Utils.CreateBotAndPoll()

	Utils.DeleteThis(bh)

	Customer.Menu(bh, db)
	Customer.Catalog(bh, db)
	Customer.ItemInfo(bh, db)
	Customer.MyCart(bh, db)
	Customer.Cabinet(bh, db)

	go func() {
		err1 := bh.Start()
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	select {}
}
