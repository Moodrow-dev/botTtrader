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
		order := Orders.NewOrder(db, user, user.ShoppingCart)
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
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Закрыть", CallbackData: "deleteThis"}}), MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Отлично! Заказ %v создан.\nДля подтверждения оплаты отправьте чек в поддержку\nИнформация о заказе в личном кабинете", order.ID)})
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
		ShowOrdersPage(0, userOrders, bot, ctx, chatID, messageID)
		return nil
	}, th.CallbackDataEqual("myOrders"))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		CDslc := strings.Split(callback.Data, " ")
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		user, _ := Users.GetByID(callback.From.ID, db)
		messageID := callback.Message.GetMessageID()
		var itemPage int64
		itemPage, err := strconv.ParseInt(CDslc[1], 10, 64)
		orders, err := Orders.GetOrdersOfCustomer(user.ID, db)
		if err != nil {
			errMsg(bot, chatID)
		}
		ShowOrdersPage(int(itemPage), orders, bot, ctx, chatID, messageID)
		return nil
	}, th.CallbackDataContains("ordPage"))
}

func OrderInfo(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		orderID, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		order, err := Orders.GetByID(int(orderID), db)
		if err != nil {
			errMsg(bot, chatID)
		}
		var btns []telego.InlineKeyboardButton
		if !order.IsPaid {
			btns = append(btns, telego.InlineKeyboardButton{Text: "Отменить заказ", CallbackData: fmt.Sprintf("deleteOrder %v", order.ID)})
		}
		btns = append(btns, telego.InlineKeyboardButton{Text: "Закрыть", CallbackData: "deleteThis"})
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyMarkup: tu.InlineKeyboard(btns), ChatID: chatID, Text: fmt.Sprintf("Заказ №%v", order.ID)})
		return nil
	}, th.CallbackDataContains("order"))
}

func DeleteOrder(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		messageID := callback.Message.GetMessageID()
		orderID, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		if err != nil {
			errMsg(bot, chatID)
			return err
		}
		Orders.Delete(int(orderID), db)
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Закрыть", CallbackData: "deleteThis"}}), MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Заказ №%v отменен", orderID)})
		return nil
	})
}

func BuyNow(bh *th.BotHandler, db *sql.DB) {
	var itemID int64
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		user, err := Users.GetByID(callback.From.ID, db)
		messageID := callback.Message.GetMessageID()
		itemID, _ = strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		if err != nil {
			errMsg(bot, chatID)
		}
		item, _ := Items.GetByID(itemID, db)
		btns := []telego.InlineKeyboardButton{{Text: "Да", CallbackData: "confirmBuy"}, {Text: "Нет", CallbackData: "deleteThis"}}
		kb := tu.InlineKeyboardGrid(tu.InlineKeyboardCols(1, btns...))
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: kb, MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("%v\nОформить заказ?", item.Name)})
		Users.Save(user, db)
		return nil
	}, th.CallbackDataContains("buyNow"))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		user, _ := Users.GetByID(callback.From.ID, db)
		item, _ := Items.GetByID(itemID, db)
		items := map[*Items.Item]int{item: 1}
		order := Orders.NewOrder(db, user, items)
		Orders.Save(order, db)
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Закрыть", CallbackData: "deleteThis"}}), MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Отлично! Заказ %v создан.\nДля подтверждения оплаты отправьте чек в поддержку\nИнформация о заказе в личном кабинете", order.ID)})
		return nil
	}, th.CallbackDataEqual("confirmBuy"))
}

func ShowOrdersPage(orderPage int, orders []*Orders.Order, bot *telego.Bot, ctx *th.Context, id telego.ChatID, messageID int) {
	backBtn := telego.InlineKeyboardButton{
		Text:         "🔙 Назад",
		CallbackData: "cabinet",
	}

	if len(orders) == 0 {
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: tu.InlineKeyboard([]telego.InlineKeyboardButton{backBtn}), MessageID: messageID, ChatID: id, Text: "Заказов пока нет"})
		return
	}

	const ordersPerPage = 5
	maxPage := (len(orders) - 1) / ordersPerPage

	if orderPage < 0 {
		orderPage = 0
	} else if orderPage > maxPage {
		orderPage = maxPage
	}

	start := orderPage * ordersPerPage
	end := start + ordersPerPage
	if end > len(orders) {
		end = len(orders)
	}
	pageOrders := orders[start:end]

	var keyboardRows [][]telego.InlineKeyboardButton

	for _, order := range pageOrders {
		btnText := fmt.Sprintf("Заказ №%v | %v ₽", order.ID, order.OrderValue)
		callbackData := fmt.Sprintf("order %v", order.ID)

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
	if orderPage > 0 {
		pgDownBtn = fmt.Sprintf("ordPage %v", orderPage-1)
	} else {
		pgDownBtn = "pageOrderErr"
	}
	navButtons = append(navButtons, telego.InlineKeyboardButton{
		Text:         "<< Назад",
		CallbackData: pgDownBtn,
	})
	var pgUpBtn string
	if orderPage < maxPage {
		pgUpBtn = fmt.Sprintf("ordPage %v", orderPage+1)
	} else {
		pgUpBtn = "pageOrderErr"
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
		Text:        fmt.Sprintf("Ваши активные заказы\nСтраница %v/%v\nВыберите заказ:", orderPage+1, maxPage+1),
		ReplyMarkup: &kb,
	})
}
