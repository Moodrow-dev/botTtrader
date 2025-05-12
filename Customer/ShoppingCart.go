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
		ShowCartPage(0, user.ShoppingCart, bot, ctx, chatID, messageID)
		return err
	}, th.CallbackDataEqual("mycart"))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		user, err := Users.GetByID(callback.From.ID, db)
		itemPage, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		if err != nil {
			errMsg(bot, chatID)
		}
		ShowCartPage(int(itemPage), user.ShoppingCart, bot, ctx, chatID, messageID)
		return nil
	}, th.CallbackDataContains("cartPage"))
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

func ShowCartPage(itemPage int, items map[*Items.Item]int, bot *telego.Bot, ctx *th.Context, id telego.ChatID, messageID int) {
	backBtn := telego.InlineKeyboardButton{
		Text:         "🔙 Назад",
		CallbackData: "customer_menu",
	}

	if len(items) == 0 {
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: tu.InlineKeyboard([]telego.InlineKeyboardButton{backBtn}), MessageID: messageID, ChatID: id, Text: "Ваша корзина пуста...\nНо это можно исправить😁"})
		return
	}

	const itemsPerPage = 5
	maxPage := (len(items) - 1) / itemsPerPage

	if itemPage < 0 {
		itemPage = 0
	} else if itemPage > maxPage {
		itemPage = maxPage
	}

	start := itemPage * itemsPerPage
	end := start + itemsPerPage
	if end > len(items) {
		end = len(items)
	}

	var keyboardRows [][]telego.InlineKeyboardButton

	for item, quan := range items {
		btnText := item.Name
		callbackData := fmt.Sprintf("item %v", item.ID)

		row := []telego.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%v - %v шт. | %v ₽", btnText, quan, item.Price),
				CallbackData: callbackData,
			},
		}
		keyboardRows = append(keyboardRows, row)
	}

	var navButtons []telego.InlineKeyboardButton
	if itemPage > 0 {
		navButtons = append(navButtons, telego.InlineKeyboardButton{
			Text:         "<< Назад",
			CallbackData: fmt.Sprintf("cartPage %v", itemPage-1),
		})
	}
	if itemPage < maxPage {
		navButtons = append(navButtons, telego.InlineKeyboardButton{
			Text:         "Вперед >>",
			CallbackData: fmt.Sprintf("cartPage %v", itemPage+1),
		})
	}

	if len(navButtons) > 0 {
		keyboardRows = append(keyboardRows, navButtons)
	}

	keyboardRows = append(keyboardRows, []telego.InlineKeyboardButton{
		{
			Text:         "Оформить заказ",
			CallbackData: "makeOrder",
		},
	})
	keyboardRows = append(keyboardRows, []telego.InlineKeyboardButton{
		{
			Text:         "Очистить корзину",
			CallbackData: "clearCart",
		},
	})

	keyboardRows = append(keyboardRows, []telego.InlineKeyboardButton{backBtn})

	kb := telego.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	_, _ = bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:      id,
		MessageID:   messageID,
		Text:        fmt.Sprintf("В вашей корзине – %v предметов\nСтраница %v/%v\nВыберите товар:", len(items), itemPage+1, maxPage+1),
		ReplyMarkup: &kb,
	})
}
