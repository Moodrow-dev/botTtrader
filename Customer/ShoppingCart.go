package Customer

import (
	"botTtrader/Items"
	"botTtrader/Orders"
	"botTtrader/Users"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"maps"
	"strconv"
	"strings"
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
			mycart = make(map[*Items.Item]int)
		}

		for item, quantity := range maps.All(mycart) {
			btns = append(btns, telego.InlineKeyboardButton{
				Text:         fmt.Sprintf("%v(%v) – %v", item.Name, quantity, float64(quantity)*item.Price),
				CallbackData: fmt.Sprintf("item %v", item.ID),
			})
		}
		if len(mycart) > 0 {
			btns = append(btns, telego.InlineKeyboardButton{
				Text:         "Оформить заказ",
				CallbackData: "makeOrder",
			})
			btns = append(btns, telego.InlineKeyboardButton{
				Text:         "Очистить корзину",
				CallbackData: "clearCart",
			})
		}
		btns = append(btns, telego.InlineKeyboardButton{
			Text:         "🔙 Назад",
			CallbackData: "customer_menu",
		})

		kb := tu.InlineKeyboard(tu.InlineKeyboardCols(1, btns...)...)
		_, err = bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ReplyMarkup: kb,
			MessageID:   messageID,
			ChatID:      chatID,
			Text:        fmt.Sprintf("В вашей корзине – %v предметов", len(mycart)),
		})
		return err
	}, th.CallbackDataEqual("mycart"))
}

func AddItemToCart(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		user, err := Users.GetByID(callback.From.ID, db)
		messageID := callback.Message.GetMessageID()
		itemID, _ := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		if err != nil {
			item, _ := Items.GetByID(itemID, db)
			user.AddToCart(item, 1)
			bot.EditMessageText(ctx, &telego.EditMessageTextParams{MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Товар: %v добавлен в корзину", item.Name)})
		}
		Users.Save(user, db)
		return nil
	}, th.CallbackDataContains("addToCart"))
}

func MakeOrder(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		user, _ := Users.GetByID(callback.From.ID, db)
		orders, _ := Orders.GetAllIDs(db)
		orderID := len(orders)
		order := Orders.NewOrder(orderID, user, user.ShoppingCart)
		Orders.Save(order, db)
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Отлично! Заказ %v создан.\nДля подтверждения оплаты отправьте чек в поддержку\nИнформация о заказе в личном кабинете", orderID)})
		return nil
	}, th.CallbackDataEqual("makeOrder"))
}

func ClearCart(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		user, _ := Users.GetByID(callback.From.ID, db)
		user.ShoppingCart = make(map[*Items.Item]int)
		kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "🔙 Назад", CallbackData: "mycart"}})
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: kb, MessageID: messageID, ChatID: chatID, Text: "Корзина успешно очищена"})
		return nil
	}, th.CallbackDataEqual("clearCart"))
}
