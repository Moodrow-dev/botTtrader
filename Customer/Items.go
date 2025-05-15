package Customer

import (
	"botTtrader/Items"
	"botTtrader/Utils"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"sort"
	"strconv"
	"strings"
)

func ItemInfo(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		itemID, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		item, err := Items.GetByID(itemID, db)
		kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Добавить в корзину", CallbackData: fmt.Sprintf("addToCart %v", itemID)}, {Text: "Купить сейчас", CallbackData: fmt.Sprintf("buyNow %v", itemID)}, {Text: "Закрыть", CallbackData: "deleteThis"}})
		if err != nil {
			ErrMsg(bot, chatID)
			return err
		}
		var quantity string
		if item.Quantity == -1 {
			quantity = "Под заказ"
		} else {
			quantity = fmt.Sprintf("%v шт.", strconv.Itoa(item.Quantity))
		}
		_, err = bot.SendPhoto(ctx, &telego.SendPhotoParams{ReplyMarkup: kb, Photo: telego.InputFile{FileID: item.PhotoId}, ChatID: chatID, Caption: fmt.Sprintf("Инфо о товаре:\n%v\nТип:%v\nОписание:\n%v\nСтоимость:%v ₽", item.Name, item.Type, item.Description, item.Price)})
		if err != nil {
			bot.SendMessage(ctx, &telego.SendMessageParams{ParseMode: telego.ModeMarkdownV2, ReplyMarkup: kb, ChatID: chatID, Text: fmt.Sprintf("*Инфо о товаре:*\n%v\n*Тип:* %v\n*Описание:*\n%v\n*Стоимость:* %v ₽\n*Осталось:* %v", Utils.EscapeMarkdown(item.Name), Utils.EscapeMarkdown(item.Type), Utils.EscapeMarkdown(item.Description), Utils.EscapeMarkdown(fmt.Sprintf("%.2f", item.Price)), Utils.EscapeMarkdown(quantity))})
		}
		return nil
	}, th.CallbackDataContains("item"))
}

func Catalog(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		CDslc := strings.Split(callback.Data, " ")
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		items, err := Items.GetAll(db)
		itemType := "Все"
		if len(CDslc) != 1 && CDslc[1] != itemType {
			itemType = CDslc[1]
			items, err = Items.GetByType(itemType, db)
		}
		if err != nil {
			ErrMsg(bot, chatID)
			return err
		}
		ShowItemsPage(0, itemType, items, bot, ctx, chatID, messageID)
		return nil
	}, th.CallbackDataContains("catalog"))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		CDslc := strings.Split(callback.Data, " ")
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		items, err := Items.GetAll(db)
		var itemPage int64
		itemType := "Все"
		if len(CDslc) == 3 {
			itemType = CDslc[1]
		}
		itemPage, err = strconv.ParseInt(CDslc[2], 10, 64)
		if err != nil {
			ErrMsg(bot, chatID)
		}
		if itemType != "Все" {
			items, err = Items.GetByType(itemType, db)
		}
		if err != nil {
			ErrMsg(bot, chatID)
		}
		ShowItemsPage(int(itemPage), itemType, items, bot, ctx, chatID, messageID)
		return nil
	}, th.CallbackDataContains("catPage"))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		items, err := Items.GetAll(db)
		if err != nil {
			ErrMsg(bot, chatID)
		}
		types := make(map[string]bool)
		typesSlice := []string{}
		for _, item := range items {
			if types[item.Type] == false {
				types[item.Type] = true
				typesSlice = append(typesSlice, item.Type)
			}
		}
		sort.Strings(typesSlice)
		typesSlice = append(typesSlice, "Все")
		btns := []telego.InlineKeyboardButton{}
		for _, itemType := range typesSlice {
			btns = append(btns, telego.InlineKeyboardButton{Text: itemType, CallbackData: fmt.Sprintf("catalog %v", itemType)})
		}

		btns = append(btns, telego.InlineKeyboardButton{
			Text:         "🔙 Назад",
			CallbackData: "catalog",
		})
		kb := tu.InlineKeyboard(tu.InlineKeyboardCols(1, btns...)...)
		bot.EditMessageText(ctx, &telego.EditMessageTextParams{ReplyMarkup: kb, MessageID: messageID, ChatID: chatID, Text: "Выберите тип товара"})
		return nil
	}, th.CallbackDataContains("itemSort"))
}

func ShowItemsPage(itemPage int, itemType string, items []*Items.Item, bot *telego.Bot, ctx *th.Context, id telego.ChatID, messageID int) {
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
