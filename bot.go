package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"log"
	"os"
)

func createBotAndPoll() (*telego.Bot, *th.BotHandler, *Owner, error) {
	owner := Owner{}
	err := godotenv.Load("settings.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	bot, err := telego.NewBot(os.Getenv("BOT_TOKEN"), telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
		return nil, nil, nil, err
	}
	upd, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	bh, _ := th.NewBotHandler(bot, upd)
	return bot, bh, &owner, nil
}

func getInfoAboutOwner(owner *Owner) {
	if owner == nil {
		//
	} else {
		//
	}
}

func addOwnerInfo(bh *th.BotHandler) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		if readOwner("owner.json") == nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: update.Message.Chat.ChatID(), Text: "Внимание!\nЧтобы пользоваться ботом далее необходимо указать свои данные."})
		}
		return nil
	}, th.AnyMessage())
}

func writeOwner(owner *Owner) {
	user, err := json.Marshal(owner)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
	os.WriteFile("owner.json", user, 0644)

}

func readOwner(name string) *Owner {
	newOwn := &Owner{}
	file, _ := os.ReadFile(name)
	err := json.Unmarshal(file, &newOwn)
	if err != nil {
		return nil
	}
	return newOwn
}
