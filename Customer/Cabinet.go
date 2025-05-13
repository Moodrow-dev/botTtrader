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
				telego.InlineKeyboardButton{Text: "✏️ Сменить имя", CallbackData: "changeName"},
			),
			tu.InlineKeyboardRow(
				telego.InlineKeyboardButton{Text: "📱 Сменить телефон", CallbackData: "changePhone"},
			),
			tu.InlineKeyboardRow(
				telego.InlineKeyboardButton{Text: "🏠 Сменить адрес", CallbackData: "changeAddress"},
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
}
