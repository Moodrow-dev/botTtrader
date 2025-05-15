package Utils

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"log"
	"os"
	"strings"
)

func CreateBotAndPoll() (*telego.Bot, *th.BotHandler, error) {
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

func DeleteThis(bh *th.BotHandler) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		bot := ctx.Bot()
		bot.DeleteMessage(ctx, &telego.DeleteMessageParams{chatID, messageID})
		return nil
	}, th.CallbackDataEqual("deleteThis"))
}

// EscapeMarkdown экранирует специальные символы MarkdownV2
func EscapeMarkdown(text string) string {
	chars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range chars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}
