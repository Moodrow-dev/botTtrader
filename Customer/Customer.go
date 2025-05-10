package Customer

import (
	"botTtrader/Items"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func Menu(bh *th.BotHandler) {
	kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Прайс", CallbackData: "price"}, {Text: "Связаться с владельцем", CallbackData: "support"}})
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		chatID := update.Message.Chat.ChatID()
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyMarkup: kb, ChatID: chatID, Text: fmt.Sprintf("Привет, %v", update.Message.From.FirstName)})
		return nil
	}, th.Or(th.CommandEqual("menu"), th.CommandEqual("start")))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		chatID := telego.ChatID{ID: update.CallbackQuery.From.ID}
		println(chatID.ID)
		bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, ReplyMarkup: kb, Text: fmt.Sprintf("Привет, %v", update.Message.From.FirstName)})
		return nil
	}, th.CallbackDataEqual("customerMenu"))
}

func Price(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		chatID := telego.ChatID{ID: update.CallbackQuery.From.ID}
		items, err := Items.GetAll(db)
		if err != nil {
			bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Произошла ошибка, обратитесь к владельцу"})
			return err
		}
		btns := []telego.InlineKeyboardButton{}
		if len(items) > 0 {
			for _, item := range items {
				btns = append(btns, telego.InlineKeyboardButton{Text: item.Name, CallbackData: fmt.Sprintf("item %v", item.ID)})
			}
			btns = append(btns, telego.InlineKeyboardButton{Text: "Назад", CallbackData: "customerMenu"})
			kb := tu.InlineKeyboard(btns)
			bot.SendMessage(ctx, &telego.SendMessageParams{ReplyMarkup: kb, Text: "Вот текущий ассортимент:", ChatID: chatID})
		} else {
			bot.SendMessage(ctx, &telego.SendMessageParams{ReplyMarkup: nil, ChatID: chatID, Text: "Пока в продаже ничего нет :("})
		}
		return nil
	}, th.CallbackDataEqual("price"))
}

func GetInfoAboutItem(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		chatID := update.Message.Chat.ChatID()
		bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Инфо о товаре"})
		return nil
	}, th.CallbackDataEqual("item"))
}
