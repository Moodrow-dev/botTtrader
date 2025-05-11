package Customer

import (
	"botTtrader/Items"
	"botTtrader/Users"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func MyCart(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		user, err := Users.GetByID(callback.From.ID, db)
		if err != nil {
			// Handle error (e.g., log it or notify the user)
			bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: chatID,
				Text:   "Ошибка при получении данных пользователя",
			})
			return err
		}

		btns := []telego.InlineKeyboardButton{}
		mycart := user.ShoppingCart
		if mycart == nil {
			// Handle empty or uninitialized cart
			mycart = []*Items.Item{} // Assuming Item is the type of cart items
		}

		for _, item := range mycart {
			btns = append(btns, telego.InlineKeyboardButton{
				Text:         fmt.Sprintf("%v – %v", item.Name, item.Price),
				CallbackData: fmt.Sprintf("item %v", item.ID),
			})
		}
		btns = append(btns, telego.InlineKeyboardButton{
			Text:         "Назад",
			CallbackData: "customer_menu",
		})

		kb := tu.InlineKeyboard(btns)
		_, err = bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ReplyMarkup: kb,
			MessageID:   messageID,
			ChatID:      chatID,
			Text:        fmt.Sprintf("В вашей корзине – %v предметов", len(mycart)),
		})
		return err
	}, th.CallbackDataEqual("mycart"))
}
