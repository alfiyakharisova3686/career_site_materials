package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Кому отправлять результаты опросов
var notifyChatIDs = []int64{
	552007058,  // Альфия
	1742778598, // Денис (разработчик)
}

// NotifyData содержит данные для уведомления о результате опроса
type NotifyData struct {
	User         *tgbotapi.User
	TestName     string
	ResultMain   string // основной результат
	ResultDetail string // дополнительная секция (необязательно)
	CompletedAt  time.Time
}

// formatUserName возвращает читаемое имя пользователя Telegram
func formatUserName(user *tgbotapi.User) string {
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if user.UserName != "" {
		return fmt.Sprintf("%s (@%s)", name, user.UserName)
	}
	return fmt.Sprintf("%s (ID: %d)", name, user.ID)
}

// sendResultNotification отправляет результат опроса в Telegram.
// Вызывается асинхронно — не блокирует работу бота.
func sendResultNotification(bot *tgbotapi.BotAPI, data NotifyData) {
	go func() {
		userName := formatUserName(data.User)
		text := buildNotificationText(data, userName)

		for _, chatID := range notifyChatIDs {
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "HTML"

			if _, err := bot.Send(msg); err != nil {
				log.Printf("[Notify] Ошибка отправки уведомления в чат %d: %v", chatID, err)
			}
		}

		log.Printf("[Notify] Результат опроса '%s' отправлен. Пользователь: %s", data.TestName, userName)
	}()
}

// buildNotificationText формирует текст уведомления для Telegram
func buildNotificationText(data NotifyData, userName string) string {
	text := fmt.Sprintf("📋 <b>Новый результат опроса</b>\n\n"+
		"<b>Тест:</b> %s\n"+
		"<b>Пользователь:</b> %s\n"+
		"<b>Telegram ID:</b> %d\n"+
		"<b>Дата:</b> %s\n\n"+
		"<b>Результат:</b>\n%s",
		data.TestName,
		userName,
		data.User.ID,
		data.CompletedAt.Format("02.01.2006 в 15:04"),
		data.ResultMain,
	)

	if data.ResultDetail != "" {
		text += fmt.Sprintf("\n\n<b>Подробнее:</b>\n%s", data.ResultDetail)
	}

	if data.User.UserName != "" {
		text += fmt.Sprintf("\n\n✉️ <a href=\"https://t.me/%s\">Написать пользователю</a>", data.User.UserName)
	}

	return text
}
