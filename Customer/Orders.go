package Customer

import (
	"botTtrader/Items"
	"botTtrader/Orders"
	"botTtrader/Users"
	"botTtrader/Utils"
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
		if callback == nil {
			return nil
		}

		// Парсим ID заказа из callback data
		orderID, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)

		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}

		// Получаем заказ из базы данных
		order, err := Orders.GetByID(int(orderID), db)
		if err != nil {
			log.Println(err)
			return err
		}

		// Формируем текст сообщения с использованием Markdown
		var msgBuilder strings.Builder
		msgBuilder.WriteString(fmt.Sprintf("*Заказ №%d*\n", order.ID))
		msgBuilder.WriteString(fmt.Sprintf("📅 *Дата*: %s\n", Utils.EscapeMarkdown(order.CreatedAt.Format("02.01.2006 15:04"))))
		msgBuilder.WriteString(fmt.Sprintf("💳 *Статус*: %s\n", GetStatusText(order.IsPaid)))
		if order.Track != "" {
			msgBuilder.WriteString(fmt.Sprintf("📦 *Трек-номер*: %s\n", order.Track))
		}
		if order.Customer != nil {
			msgBuilder.WriteString(fmt.Sprintf("👤 *Клиент*: %s\n", order.Customer.Name))
		}
		msgBuilder.WriteString("\n*Товары*:\n")

		// Перечисляем товары из мапы
		itemIndex := 1
		for item, quantity := range order.Items {
			msgBuilder.WriteString(
				Utils.EscapeMarkdown(fmt.Sprintf(
					"%d. %s\n   Кол-во: %d x %.2f ₽ = %.2f ₽\n",
					itemIndex, item.Name, quantity, item.Price, float64(quantity)*item.Price,
				)),
			)
			itemIndex++
		}

		// Добавляем итоговую сумму
		msgBuilder.WriteString("\n*Итого*:" + Utils.EscapeMarkdown(fmt.Sprintf(" %.2f ₽\n", order.OrderValue)))

		// Формируем кнопки
		var btns []telego.InlineKeyboardButton
		if !order.IsPaid {
			btns = append(btns, telego.InlineKeyboardButton{
				Text:         "Отменить заказ",
				CallbackData: fmt.Sprintf("deleteOrder %d", order.ID),
			})
		}
		btns = append(btns, telego.InlineKeyboardButton{
			Text:         "Закрыть",
			CallbackData: "deleteThis",
		})

		// Отправляем сообщение
		bot.SendMessage(ctx, &telego.SendMessageParams{ParseMode: telego.ModeMarkdownV2, ReplyMarkup: tu.InlineKeyboard(btns), ChatID: chatID, Text: msgBuilder.String()})

		if err != nil {
			return err
		}

		return nil
	}, th.CallbackDataContains("order"))
}

// Вспомогательная функция для получения текста статуса
func GetStatusText(isPaid bool) string {
	if isPaid {
		return "Оплачен"
	}
	return "Ожидает оплаты"
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
