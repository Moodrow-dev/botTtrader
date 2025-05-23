package main

import (
	"botTtrader/Customer"
	"botTtrader/Items"
	"botTtrader/Orders"
	"botTtrader/Users"
	"botTtrader/Utils"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
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

	itemTypes := []string{"Pin", "Sticker", "Doll"}
	for i := range int64(15) {
		err = Items.Save(Items.Item{i, itemTypes[rand.Intn(3)], fmt.Sprintf("Товар %v", i), int(i), "Пробный товар", 100 * float64(i), "123", time.Now()}, db)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, bh, _ := Utils.CreateBotAndPoll()

	Utils.DeleteThis(bh)

	SetupHandlers(bh, db)
	Customer.StartForCustomer(bh, db)

	go func() {
		err1 := bh.Start()
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	select {}
}
