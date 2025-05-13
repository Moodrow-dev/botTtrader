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
	"log"
	"strconv"
	"strings"
)

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
		err := Orders.Save(order, db)
		if err != nil {
			log.Println(err)
			return err
		}
		user.ShoppingCart = make(map[*Items.Item]int)
		err = Users.Save(user, db)
		if err != nil {
			log.Println(err)
			return err
		}
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Отлично! Заказ %v создан.\nДля подтверждения оплаты отправьте чек в поддержку\nИнформация о заказе в личном кабинете", orderID)})
		return nil
	}, th.CallbackDataEqual("makeOrder"))
}

func MyOrders(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		user, _ := Users.GetByID(callback.From.ID, db)
		userOrders, _ := Orders.GetOrdersOfCustomer(user.ID, db)
		btns := []telego.InlineKeyboardButton{}
		for _, order := range userOrders {
			btns = append(btns, telego.InlineKeyboardButton{Text: fmt.Sprintf("Заказ №%v", order.ID), CallbackData: fmt.Sprintf("order %v", order.ID)})
		}
		btns = append(btns, telego.InlineKeyboardButton{Text: "🔙 Назад", CallbackData: "cabinet"})

		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: tu.InlineKeyboard(btns), MessageID: messageID, ChatID: chatID, Text: "Ваши активные заказы"})
		return nil
	}, th.CallbackDataEqual("myOrders"))
}

func BuyNow(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		user, err := Users.GetByID(callback.From.ID, db)
		messageID := callback.Message.GetMessageID()
		itemID, _ := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		if err != nil {
			errMsg(bot, chatID)
		}
		item, _ := Items.GetByID(itemID, db)
		user.AddToCart(item, 1)
		kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Закрыть", CallbackData: "deleteThis"}})
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: kb, MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Оформить заказ?", item.Name)})
		Users.Save(user, db)
		return nil
	}, th.CallbackDataContains("buyNow"))
}

func ShowOrdersPage(itemPage int, itemType string, items []*Items.Item, bot *telego.Bot, ctx *th.Context, id telego.ChatID, messageID int) {
	backBtn := telego.InlineKeyboardButton{
		Text:         "🔙 Назад",
		CallbackData: "customerMenu",
	}

	if len(items) == 0 {
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: tu.InlineKeyboard([]telego.InlineKeyboardButton{backBtn}), MessageID: messageID, ChatID: id, Text: "Товаров пока нет"})
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
	pageItems := items[start:end]

	var keyboardRows [][]telego.InlineKeyboardButton

	for _, item := range pageItems {
		btnText := fmt.Sprintf("%v | %v ₽", item.Name, item.Price)
		callbackData := fmt.Sprintf("item %v", item.ID)
		if item.Quantity == 0 {
			btnText = "Нет в наличии"
			callbackData = "notAvailableItem"
		}

		row := []telego.InlineKeyboardButton{
			{
				Text:         btnText,
				CallbackData: callbackData,
			},
		}
		keyboardRows = append(keyboardRows, row)
	}

	var navButtons []telego.InlineKeyboardButton
	var pgDownBtn string
	if itemPage > 0 {
		pgDownBtn = fmt.Sprintf("catPage %v %v", itemType, itemPage-1)
	} else {
		pgDownBtn = "pageItemErr"
	}
	navButtons = append(navButtons, telego.InlineKeyboardButton{
		Text:         "<< Назад",
		CallbackData: pgDownBtn,
	})
	navButtons = append(navButtons, telego.InlineKeyboardButton{
		Text:         "Фильтр",
		CallbackData: "itemSort",
	})
	var pgUpBtn string
	if itemPage < maxPage {
		pgUpBtn = fmt.Sprintf("catPage %v %v", itemType, itemPage+1)
	} else {
		pgUpBtn = "pageItemErr"
	}
	navButtons = append(navButtons, telego.InlineKeyboardButton{
		Text:         "Вперед >>",
		CallbackData: pgUpBtn,
	})

	if len(navButtons) > 0 {
		keyboardRows = append(keyboardRows, navButtons)
	}

	keyboardRows = append(keyboardRows, []telego.InlineKeyboardButton{backBtn})

	kb := telego.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	_, _ = bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:      id,
		MessageID:   messageID,
		Text:        fmt.Sprintf("Тип: %v\nСтраница %v/%v\nВыберите товар:", itemType, itemPage+1, maxPage+1),
		ReplyMarkup: &kb,
	})
}
