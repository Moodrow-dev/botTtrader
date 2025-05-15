package main

import (
	"botTtrader/Items"
	"botTtrader/Users"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Константы для состояний
const (
	stateNormal byte = iota
	stateEditingName
	stateEditingPhone
	stateEditingAddress
	stateEditingQuantity
)

// userState хранит состояние пользователя
type userState struct {
	state  byte
	itemID int64
	msgID  int
}

// SetupHandlers настраивает все обработчики
func SetupHandlers(bh *th.BotHandler, db *sql.DB) {
	var userStates = make(map[int64]userState)
	var mu sync.Mutex

	// Изменение имени
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		if callback == nil || callback.Message == nil {
			return nil
		}
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		userID := callback.From.ID

		_, err := bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      "Введите ваше ФИО (например: Иванов Иван Иванович):",
		})

		if err != nil {
			log.Printf("Failed to request name change: %v", err)
			return err
		}

		mu.Lock()
		userStates[userID] = userState{state: stateEditingName}
		mu.Unlock()
		return nil
	}, th.CallbackDataEqual("changeName"))

	// Изменение телефона
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		if callback == nil || callback.Message == nil {
			return nil
		}
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		userID := callback.From.ID

		_, err := bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      "Введите ваш номер телефона (можете поделиться контактом) (например: +79123456789):",
		})

		if err != nil {
			log.Printf("Failed to request phone change: %v", err)
			return err
		}

		mu.Lock()
		userStates[userID] = userState{state: stateEditingPhone}
		mu.Unlock()
		return nil
	}, th.CallbackDataEqual("changePhone"))

	// Изменение адреса
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		if callback == nil || callback.Message == nil {
			return nil
		}
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		userID := callback.From.ID

		_, err := bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      "Введите ваш адрес доставки (например: г. Москва, ул. Ленина, д. 10, кв. 25):",
		})

		if err != nil {
			log.Printf("Failed to request address change: %v", err)
			return err
		}

		mu.Lock()
		userStates[userID] = userState{state: stateEditingAddress}
		mu.Unlock()
		return nil
	}, th.CallbackDataEqual("changeAddress"))

	// Изменение количества товара
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		bot := ctx.Bot()
		callback := update.CallbackQuery
		if callback == nil || callback.Message == nil {
			return nil
		}
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()
		userID := callback.From.ID

		itemID, err := strconv.ParseInt(strings.Split(callback.Data, " ")[1], 10, 64)
		item, err := Items.GetByID(itemID, db)
		if err != nil {
			log.Printf("Invalid item ID: %v", err)
			return err
		}

		mu.Lock()
		userStates[userID] = userState{
			state:  stateEditingQuantity,
			itemID: itemID,
			msgID:  messageID,
		}
		mu.Unlock()

		var quan int
		if item.Quantity != -1 {
			quan = item.Quantity
		} else {
			quan = 1000
		}
		_, err = bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      fmt.Sprintf("Введите нужное количество товара (макс. %v):", quan),
		})

		if err != nil {
			log.Printf("Failed to edit message: %v", err)
			return err
		}
		return nil
	}, th.CallbackDataContains("changeQuantity"))

	// Общий обработчик сообщений
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		if update.Message == nil {
			return nil
		}

		bot := ctx.Bot()
		userID := update.Message.From.ID
		chatID := telego.ChatID{ID: update.Message.Chat.ID}
		text := update.Message.Text
		contact := update.Message.Contact

		mu.Lock()
		stateData, ok := userStates[userID]
		if !ok || stateData.state == stateNormal {
			mu.Unlock()
			return nil
		}
		mu.Unlock()

		user, err := Users.GetByID(userID, db)
		if err != nil {
			log.Printf("Failed to get user: %v", err)
			return err
		}

		var successMessage string
		var errorMessage string
		var replyMarkup *telego.InlineKeyboardMarkup

		switch stateData.state {
		case stateEditingPhone:
			phoneRegex := regexp.MustCompile(`^\+7\d{10}$|^8\d{10}$`)
			var phone string
			if contact != nil {
				phone = "+" + contact.PhoneNumber
			}
			if (text != "" && phoneRegex.MatchString(text)) || (phone != "" && phoneRegex.MatchString(phone)) {
				user.Phone = phone
				if text != "" {
					user.Phone = text
				}
				successMessage = "✅ Номер телефона успешно обновлен!"
				replyMarkup = tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						telego.InlineKeyboardButton{Text: "Вернуться в кабинет", CallbackData: "cabinet"},
					),
				)
			} else {
				errorMessage = "⚠️ Пожалуйста, введите корректный номер телефона (например: +79123456789)"
			}
		case stateEditingAddress:
			if len(text) > 20 && strings.Contains(text, ",") {
				user.Address = text
				successMessage = "✅ Адрес успешно обновлен!"
				replyMarkup = tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						telego.InlineKeyboardButton{Text: "Вернуться в кабинет", CallbackData: "cabinet"},
					),
				)
			} else {
				errorMessage = "⚠️ Пожалуйста, введите полный адрес (например: г. Москва, ул. Ленина, д. 10, кв. 25)"
			}
		case stateEditingName:
			if len(text) > 5 && strings.Contains(text, " ") {
				user.Name = text
				successMessage = "✅ ФИО успешно обновлено!"
				replyMarkup = tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						telego.InlineKeyboardButton{Text: "Вернуться в кабинет", CallbackData: "cabinet"},
					),
				)
			} else {
				errorMessage = "⚠️ Пожалуйста, введите полное ФИО (например: Иванов Иван Иванович)"
			}
		case stateEditingQuantity:
			newQuantity, err := strconv.ParseInt(text, 10, 64)
			if err != nil || newQuantity <= 0 {
				errorMessage = "⚠️ Пожалуйста, введите корректное количество (целое число больше 0)"
			} else {
				for item, quantity := range user.ShoppingCart {
					if item.ID == stateData.itemID {
						newItem, err := Items.GetByID(item.ID, db)
						if err != nil {
							errorMessage = fmt.Sprintf("Ошибка при получении товара: %v", err)
							break
						}
						if quantity != -1 && newItem.Quantity < int(newQuantity) {
							errorMessage = fmt.Sprintf("⚠️ Такого количества товара нет в наличии. Осталось %v штук", newItem.Quantity)
						} else {
							user.ShoppingCart[item] = int(newQuantity) // Обновляем количество в мапе
						}
						break
					}
				}
				successMessage = fmt.Sprintf("✅ Количество товара изменено на %d!", newQuantity)
				replyMarkup = tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						telego.InlineKeyboardButton{Text: "Вернуться в корзину", CallbackData: "myCart"},
					),
				)
			}
		}

		if errorMessage != "" {
			_, err = bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: chatID,
				Text:   errorMessage,
			})
			return err
		}

		if err := Users.Save(user, db); err != nil {
			log.Printf("Failed to save user: %v", err)
			_, err = bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: chatID,
				Text:   "⚠️ Произошла ошибка при сохранении данных",
			})
			return err
		}

		mu.Lock()
		delete(userStates, userID)
		mu.Unlock()

		_, err = bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID:      chatID,
			Text:        successMessage,
			ReplyMarkup: replyMarkup,
		})
		return err
	}, th.AnyMessage())
}
