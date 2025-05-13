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
	"sort"
	"strconv"
	"strings"
)

type Pair struct {
	Key   *Items.Item
	Value int
}

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
			errMsg(bot, chatID)
		}
		item, _ := Items.GetByID(itemID, db)
		user.AddToCart(item, 1)
		kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Закрыть", CallbackData: "deleteThis"}})
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: kb, MessageID: messageID, ChatID: chatID, Text: fmt.Sprintf("Товар: %v добавлен в корзину", item.Name)})
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
		Users.Save(user, db)
		kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "🔙 Назад", CallbackData: "mycart"}})
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: kb, MessageID: messageID, ChatID: chatID, Text: "Корзина успешно очищена"})
		return nil
	}, th.CallbackDataEqual("clearCart"))
}

func CartItem(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		itemID, _ := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		user, _ := Users.GetByID(callback.From.ID, db)
		item, _ := Items.GetByID(itemID, db)
		quantity := 0
		for existingItem, quan := range user.ShoppingCart {
			if existingItem.ID == item.ID {
				quantity = quan
			}
		}
		btns := []telego.InlineKeyboardButton{
			{Text: "Указать количество", CallbackData: fmt.Sprintf("changeQuantity %v", item.ID)},
			{Text: "Удалить товар", CallbackData: fmt.Sprintf("deleteItemInCart %v", item.ID)},
			{Text: "Закрыть", CallbackData: "deleteThis"},
		}
		kb := tu.InlineKeyboard(btns)
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyMarkup: kb, ChatID: chatID, Text: fmt.Sprintf("Инфо о товаре:\n%v\nТип:%v\nОписание:\n%v\nСтоимость: %v ₽\nВ корзине:%v шт.", item.Name, item.Type, item.Description, item.Price, quantity)})
		return nil
	}, th.CallbackDataContains("cartItem"))
}

func ShowCartPage(itemPage int, items map[*Items.Item]int, bot *telego.Bot, ctx *th.Context, id telego.ChatID, messageID int) {
	backBtn := telego.InlineKeyboardButton{
		Text:         "🔙 Назад",
		CallbackData: "customer_menu",
	}

	if len(items) == 0 {
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ReplyMarkup: tu.InlineKeyboard([]telego.InlineKeyboardButton{backBtn}),
			MessageID:   messageID,
			ChatID:      id,
			Text:        "Ваша корзина пуста...\nНо это можно исправить😁",
		})
		return
	}

	const itemsPerPage = 5

	// Конвертируем мапу в срез пар и сортируем по ID
	type Pair struct {
		Key   *Items.Item
		Value int
	}

	allPairs := make([]Pair, 0, len(items))
	for item, quantity := range items {
		allPairs = append(allPairs, Pair{item, quantity})
	}

	// Сортируем по возрастанию ID
	sort.Slice(allPairs, func(i, j int) bool {
		return allPairs[i].Key.ID < allPairs[j].Key.ID
	})

	// Рассчитываем количество страниц
	maxPage := (len(allPairs) - 1) / itemsPerPage

	// Корректируем номер страницы
	if itemPage < 0 {
		itemPage = 0
	} else if itemPage > maxPage {
		itemPage = maxPage
	}

	// Получаем элементы для текущей страницы
	start := itemPage * itemsPerPage
	end := start + itemsPerPage
	if end > len(allPairs) {
		end = len(allPairs)
	}
	pagePairs := allPairs[start:end]

	// Создаем кнопки для товаров
	var keyboardRows [][]telego.InlineKeyboardButton
	for _, pair := range pagePairs {
		item := pair.Key
		quantity := pair.Value
		row := []telego.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%s - %d шт. | %d ₽", item.Name, quantity, int(item.Price)),
				CallbackData: fmt.Sprintf("cartItem %v", item.ID),
			},
		}
		keyboardRows = append(keyboardRows, row)
	}

	var navButtons []telego.InlineKeyboardButton
	var pgDownBtn string
	if itemPage > 0 {
		pgDownBtn = fmt.Sprintf("cartPage %v", itemPage-1)
	} else {
		pgDownBtn = "pageItemErr"
	}
	navButtons = append(navButtons, telego.InlineKeyboardButton{
		Text:         "<< Назад",
		CallbackData: pgDownBtn,
	})
	var pgUpBtn string
	if itemPage < maxPage {
		pgUpBtn = fmt.Sprintf("cartPage %v", itemPage+1)
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

	keyboardRows = append(keyboardRows, []telego.InlineKeyboardButton{
		{Text: "Оформить заказ", CallbackData: "makeOrder"},
	})
	keyboardRows = append(keyboardRows, []telego.InlineKeyboardButton{
		{Text: "Очистить корзину", CallbackData: "clearCart"},
	})
	keyboardRows = append(keyboardRows, []telego.InlineKeyboardButton{backBtn})

	_, _ = bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:      id,
		MessageID:   messageID,
		Text:        fmt.Sprintf("В вашей корзине – %d предметов\nСтраница %d/%d\nВыберите товар:", len(items), itemPage+1, maxPage+1),
		ReplyMarkup: &telego.InlineKeyboardMarkup{InlineKeyboard: keyboardRows},
	})
}
