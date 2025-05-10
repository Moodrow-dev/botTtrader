package main

import (
	"botTtrader/Users"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"log"
	"os"
)

func createBotAndPoll() (*telego.Bot, *th.BotHandler, error) {
	err := godotenv.Load("settings.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	bot, err := telego.NewBot(os.Getenv("BOT_TOKEN"), telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	upd, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	bh, _ := th.NewBotHandler(bot, upd)
	return bot, bh, nil
}

//func getInfoAboutOwner(db *sql.DB) {
//	owner, err := Users.GetOwner(db)
//	if err == nil {
//		//
//	} else {
//		//
//	}
//}

func addOwnerInfo(bh *th.BotHandler) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		if readOwner("owner.json") == nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: update.Message.Chat.ChatID(), Text: "Внимание!\nЧтобы пользоваться ботом далее необходимо указать свои данные."})
		}
		return nil
	}, th.AnyMessage())
}

func writeOwner(owner *Users.User) {
	user, err := json.Marshal(owner)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
	os.WriteFile("owner.json", user, 0644)

}

func readOwner(name string) *Users.User {
	newOwn := &Users.User{}
	file, _ := os.ReadFile(name)
	err := json.Unmarshal(file, &newOwn)
	if err != nil {
		return nil
	}
	return newOwn
}
