package Customer

import (
	"botTtrader/Users"
	"context"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func errMsg(bot *telego.Bot, id telego.ChatID) {
	errKb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Закрыть", CallbackData: "deleteThis"}})
	bot.SendMessage(context.Background(), &telego.SendMessageParams{ReplyMarkup: errKb, ChatID: id, Text: "Произошла ошибка, обратитесь к владельцу"})
}

func Menu(bh *th.BotHandler, db *sql.DB) {
	kb := tu.InlineKeyboard(tu.InlineKeyboardCols(3, []telego.InlineKeyboardButton{
		{Text: "🛍 Товары", CallbackData: "`catalog`"},
		{Text: "🛒 Корзина", CallbackData: "mycart"},
		{Text: "📇 Личный кабинет", CallbackData: "cabinet"},
		{Text: "📱 Поддержка", CallbackData: "support"}}...)...)
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		chatID := update.Message.Chat.ChatID()
		user := update.Message.From
		_, err := Users.GetByID(user.ID, db)
		if err != nil {
			Users.Save(Users.NewUser(user, false), db)
		}
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyMarkup: kb, ChatID: chatID, Text: fmt.Sprintf("Привет, %v", update.Message.From.FirstName)})
		return nil
	}, th.Or(th.CommandEqual("menu"), th.CommandEqual("start")))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery

		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		user := &telego.User{
			ID:        callback.Message.GetChat().ID,
			FirstName: callback.Message.GetChat().FirstName,
			LastName:  callback.Message.GetChat().LastName,
			Username:  callback.Message.GetChat().Username,
		}
		_, err := Users.GetByID(user.ID, db)
		if err != nil {
			Users.Save(Users.NewUser(user, false), db)
		}

		bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   callback.Message.GetMessageID(),
			Text:        fmt.Sprintf("Привет, %s", callback.From.FirstName),
			ReplyMarkup: kb,
		})
		return nil
	}, th.CallbackDataEqual("customer_menu"))
}
