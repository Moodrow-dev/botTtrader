package Customer

import (
	"botTtrader/Items"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"strconv"
	"strings"
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
		callback := update.CallbackQuery

		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}

		_, err := bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   callback.Message.GetMessageID(),
			Text:        fmt.Sprintf("Привет, %s", callback.From.FirstName),
			ReplyMarkup: kb,
		})

		if err != nil {
			return err
		}

		return nil
	}, th.CallbackDataEqual("customerMenu"))
}

func Price(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
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
			bot.EditMessageText(ctx, &telego.EditMessageTextParams{MessageID: messageID, ReplyMarkup: kb, Text: "Вот текущий ассортимент:", ChatID: chatID})
		} else {
			bot.EditMessageText(ctx, &telego.EditMessageTextParams{MessageID: messageID, ChatID: chatID, Text: "Пока в продаже ничего нет :("})
		}
		return nil
	}, th.CallbackDataEqual("price"))
}

func GetInfoAboutItem(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		itemID, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		item, err := Items.GetByID(int(itemID), db)
		if err != nil {
			bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Произошла ошибка, обратитесь к владельцу"})
			return err
		}
		_, err = bot.SendPhoto(ctx, &telego.SendPhotoParams{Photo: telego.InputFile{FileID: item.PhotoId}, ChatID: chatID, Caption: fmt.Sprintf("Инфо о товаре:\n%v\nТип:%v\nОписание:\n%v\nСтоимость:%v ₽", item.Name, item.Type, item.Description, item.Price)})
		if err != nil {
			bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: fmt.Sprintf("Инфо о товаре:\n%v\nТип:%v\nОписание:\n%v\nСтоимость:%v ₽", item.Name, item.Type, item.Description, item.Price)})
		}
		return nil
	}, th.CallbackDataContains("item"))
}
