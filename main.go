package main

import "log"

func main() {
	_, bh, _, _ := createBotAndPoll()

	addOwnerInfo(bh)

	go func() {
		err1 := bh.Start()
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	select {}
}
