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

func ItemInfo(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		itemID, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		item, err := Items.GetByID(itemID, db)
		kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{{Text: "Добавить в корзину", CallbackData: fmt.Sprintf("addToCart %v", itemID)}, {Text: "Купить сейчас", CallbackData: fmt.Sprintf("buyNow %v", itemID)}, {Text: "Закрыть", CallbackData: "deleteThis"}})
		if err != nil {
			errMsg(bot, chatID)
			return err
		}
		_, err = bot.SendPhoto(ctx, &telego.SendPhotoParams{ReplyMarkup: kb, Photo: telego.InputFile{FileID: item.PhotoId}, ChatID: chatID, Caption: fmt.Sprintf("Инфо о товаре:\n%v\nТип:%v\nОписание:\n%v\nСтоимость:%v ₽", item.Name, item.Type, item.Description, item.Price)})
		if err != nil {
			bot.SendMessage(ctx, &telego.SendMessageParams{ReplyMarkup: kb, ChatID: chatID, Text: fmt.Sprintf("Инфо о товаре:\n%v\nТип:%v\nОписание:\n%v\nСтоимость:%v ₽", item.Name, item.Type, item.Description, item.Price)})
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
		var items []*Items.Item
		var itemType string
		var err error
		if len(CDslc) == 2 {
			itemType = CDslc[1]
			items, err = Items.GetByType(itemType, db)
		} else {
			items, err := Items.GetAll(db)
		}
		if err != nil {
			errMsg(bot, chatID)
			return err
		}
		ShowPage(0, items, bot, ctx, chatID, messageID)
		return nil
	}, th.CallbackDataContains("catalog"))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		items, err := Items.GetAll(db)
		itemPage, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		if err != nil {
			errMsg(bot, chatID)
		}
		ShowPage(int(itemPage), items, bot, ctx, chatID, messageID)
		return nil
	}, th.CallbackDataContains("catPage"))

	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		return nil
	}, th.CallbackDataContains("itemSortBy"))
}

func ShowPage(itemPage int, items []*Items.Item, bot *telego.Bot, ctx *th.Context, id telego.ChatID, messageID int) {
	backBtn := telego.InlineKeyboardButton{
		Text:         "🔙 Назад",
		CallbackData: "customer_menu",
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
		btnText := item.Name
		callbackData := fmt.Sprintf("item %v", item.ID)
		if item.Quantity == 0 {
			btnText = "Нет в наличии"
			callbackData = ""
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
	if itemPage > 0 {
		navButtons = append(navButtons, telego.InlineKeyboardButton{
			Text:         "<< Назад",
			CallbackData: fmt.Sprintf("catPage %v", itemPage-1),
		})
	}
	navButtons = append(navButtons, telego.InlineKeyboardButton{
		Text:         "Фильтр",
		CallbackData: fmt.Sprintf("itemSort", itemPage),
	})
	if itemPage < maxPage {
		navButtons = append(navButtons, telego.InlineKeyboardButton{
			Text:         "Вперед >>",
			CallbackData: fmt.Sprintf("catPage %v", itemPage+1),
		})
	}

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
		Text:        fmt.Sprintf("Страница %v/%v\nВыберите товар:", itemPage+1, maxPage+1),
		ReplyMarkup: &kb,
	})
}
