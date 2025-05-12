package Customer

import (
	"botTtrader/Users"
	"botTtrader/Utils"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"log"
	"regexp"
	"strings"
)

// Состояния для машины состояний редактирования профиля
const (
	stateNormal = iota
	stateEditingName
	stateEditingPhone
	stateEditingAddress
)

// Cabinet настройки личного кабинета
func Cabinet(bh *th.BotHandler, db *sql.DB) {
	if bh == nil || db == nil {
		log.Println("Nil handler or db in Cabinet")
		return
	}

	// Главное меню кабинета
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		if update.CallbackQuery == nil || update.CallbackQuery.Message == nil {
			log.Println("Nil CallbackQuery, Message or From in Cabinet handler")
			return nil
		}

		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()

		user, err := Users.GetByID(callback.From.ID, db)
		if err != nil {
			log.Printf("Failed to get user: %v", err)
			return err
		}

		// Форматируем текст с markdown
		text := fmt.Sprintf(
			"*Ваш аккаунт:*\n\n"+
				"*ФИО:* %s\n"+
				"*Username:* @%s\n"+
				"*Скидка:* %d%%\n"+
				"*Телефон:* %s\n"+
				"*Адрес:* `%s`\n"+
				"*ID:* `%d`\n\n"+
				"Для заказов нужны: ФИО, Адрес и Телефон",
			Utils.EscapeMarkdown(user.Name),
			Utils.EscapeMarkdown(user.Account),
			user.Discount,
			Utils.EscapeMarkdown(user.Phone),
			Utils.EscapeMarkdown(user.Address),
			user.ID,
		)
		// Клавиатура кабинета
		kb := tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				telego.InlineKeyboardButton{Text: "✏️ Сменить имя", CallbackData: "change_name"},
			),
			tu.InlineKeyboardRow(
				telego.InlineKeyboardButton{Text: "📱 Сменить телефон", CallbackData: "change_phone"},
			),
			tu.InlineKeyboardRow(
				telego.InlineKeyboardButton{Text: "🏠 Сменить адрес", CallbackData: "change_address"},
			),
			tu.InlineKeyboardRow(
				telego.InlineKeyboardButton{Text: "📦 Мои заказы", CallbackData: "my_orders"},
			),
			tu.InlineKeyboardRow(
				telego.InlineKeyboardButton{Text: "🔙 Назад", CallbackData: "customer_menu"},
			),
		)

		_, err = bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        text,
			ParseMode:   telego.ModeMarkdownV2,
			ReplyMarkup: kb,
		})

		if err != nil {
			log.Printf("Failed to edit cabinet message: %v", err)
		}

		return nil
	}, th.CallbackDataEqual("cabinet"))

	// Обработчики изменения данных
	SetupEditHandlers(bh, db)
}

// SetupEditHandlers настраивает обработчики для редактирования профиля
func SetupEditHandlers(bh *th.BotHandler, db *sql.DB) {
	var state byte
	// Изменение имени
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// Проверки на nil аналогично Cabinet
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()

		// Запрашиваем новое имя
		_, err := bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      "Введите ваше ФИО (например: Иванов Иван Иванович):",
		})

		if err != nil {
			log.Printf("Failed to request name change: %v", err)
		}
		state = stateEditingName
		// Устанавливаем состояние "редактирование имени"
		// Здесь можно использовать map или БД для хранения состояний
		return nil
	}, th.CallbackDataEqual("change_name"))

	// Изменение телефона
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// Аналогичные проверки
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()

		// Запрашиваем новый телефон
		_, err := bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      "Введите ваш номер телефона(можете поделиться контактом) (например: +79123456789):",
		})

		if err != nil {
			log.Printf("Failed to request phone change: %v", err)
		}

		state = stateEditingPhone
		// Устанавливаем состояние "редактирование телефона"
		return nil
	}, th.CallbackDataEqual("change_phone"))

	// Изменение адреса
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		// Аналогичные проверки
		bot := ctx.Bot()
		callback := update.CallbackQuery
		chatID := telego.ChatID{ID: callback.Message.GetChat().ID}
		messageID := callback.Message.GetMessageID()

		// Запрашиваем новый адрес
		_, err := bot.EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      "Введите ваш адрес доставки (например: г. Москва, ул. Ленина, д. 10, кв. 25):",
		})

		if err != nil {
			log.Printf("Failed to request address change: %v", err)
		}
		state = stateEditingAddress
		// Устанавливаем состояние "редактирование адреса"
		return nil
	}, th.CallbackDataEqual("change_address"))

	// Обработка введенных данных
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		if update.Message == nil || update.Message.From == nil {
			return nil
		}

		bot := ctx.Bot()
		userID := update.Message.From.ID
		chatID := telego.ChatID{ID: update.Message.Chat.ID}
		text := update.Message.Text
		contact := update.Message.Contact
		var phone string
		if contact != nil {
			phone = "+" + contact.PhoneNumber
		}

		// Получаем пользователя
		user, err := Users.GetByID(userID, db)
		if err != nil {
			log.Printf("Failed to get user: %v", err)
			return err
		}

		var successMessage string

		if state == stateEditingPhone {
			if strings.Contains(text, "+7") || regexp.MustCompile(`\d{11}`).MatchString(text) {
				// Обновляем телефон
				user.Phone = text
			} else if phone != "" {
				user.Phone = phone
			}
			successMessage = "✅ Номер телефона успешно обновлен!"
			state = stateNormal
		} else if state == stateEditingAddress && len(text) > 20 && strings.Contains(text, ",") {
			// Обновляем адрес
			user.Address = text
			successMessage = "✅ Адрес успешно обновлен!"
			state = stateNormal
		} else if state == stateEditingName {
			// Обновляем имя
			user.Name = text
			successMessage = "✅ ФИО успешно обновлено!"
			state = stateNormal
		}

		// Сохраняем изменения
		if err := Users.Save(user, db); err != nil {
			log.Printf("Failed to save user: %v", err)
			_, _ = bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: chatID,
				Text:   "⚠️ Произошла ошибка при сохранении данных",
			})
			return nil
		}

		if err != nil {
			log.Printf("Failed to send success message: %v", err)
		}
		// Возвращаем в кабинет
		_, err = bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: chatID,
			Text:   successMessage,
			ReplyMarkup: tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					telego.InlineKeyboardButton{Text: "Вернуться в кабинет", CallbackData: "cabinet"},
				),
			),
		})

		return nil
	}, th.AnyMessage())
}
