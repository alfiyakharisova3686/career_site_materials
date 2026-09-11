package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// EmailData содержит данные для письма с результатом опроса
type EmailData struct {
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

// newlineToBreak заменяет переносы строк на HTML <br>
func newlineToBreak(s string) string {
	return strings.ReplaceAll(s, "\n", "<br>")
}

// sendResultEmail отправляет письмо с результатом опроса заказчику.
// Вызывается асинхронно — не блокирует работу бота.
func sendResultEmail(data EmailData) {
	go func() {
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := os.Getenv("SMTP_PORT")
		smtpUser := os.Getenv("SMTP_USER")
		smtpPass := os.Getenv("SMTP_PASS")
		notifyEmail := os.Getenv("NOTIFY_EMAIL")

		if smtpHost == "" || smtpUser == "" || smtpPass == "" || notifyEmail == "" {
			log.Println("[Email] SMTP не настроен в .env, отправка пропущена")
			return
		}

		userName := formatUserName(data.User)
		subject := fmt.Sprintf("Новый результат: %s — %s", data.TestName, userName)
		body := buildEmailHTML(data, userName)

		msg := "From: " + smtpUser + "\r\n" +
			"To: " + notifyEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			body

		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		addr := smtpHost + ":" + smtpPort

		if err := smtp.SendMail(addr, auth, smtpUser, []string{notifyEmail}, []byte(msg)); err != nil {
			log.Printf("[Email] Ошибка отправки для пользователя %s: %v", userName, err)
			return
		}

		log.Printf("[Email] Результат опроса '%s' отправлен. Пользователь: %s", data.TestName, userName)
	}()
}

// buildEmailHTML формирует HTML-тело письма
func buildEmailHTML(data EmailData, userName string) string {
	detailBlock := ""
	if data.ResultDetail != "" {
		detailBlock = fmt.Sprintf(`
		<h3 style="color:#4a4a8a;margin-top:24px;">Что проявляется сильнее всего</h3>
		<div style="background:#f0f4ff;padding:15px 18px;border-left:4px solid #7b7fd4;border-radius:4px;">
			%s
		</div>`, newlineToBreak(data.ResultDetail))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#f4f4f4;font-family:Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="padding:30px 0;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.1);">

        <tr><td style="background:#4a4a8a;padding:24px 30px;">
          <h1 style="margin:0;color:#ffffff;font-size:20px;">📋 Новый результат опроса</h1>
        </td></tr>

        <tr><td style="padding:24px 30px;">
          <table width="100%%" cellpadding="6" cellspacing="0" style="margin-bottom:20px;border-collapse:collapse;">
            <tr style="background:#f5f5f5;">
              <td style="font-weight:bold;width:140px;padding:8px 10px;">Тест</td>
              <td style="padding:8px 10px;">%s</td>
            </tr>
            <tr>
              <td style="font-weight:bold;padding:8px 10px;">Пользователь</td>
              <td style="padding:8px 10px;">%s</td>
            </tr>
            <tr style="background:#f5f5f5;">
              <td style="font-weight:bold;padding:8px 10px;">Telegram ID</td>
              <td style="padding:8px 10px;">%d</td>
            </tr>
            <tr>
              <td style="font-weight:bold;padding:8px 10px;">Дата и время</td>
              <td style="padding:8px 10px;">%s</td>
            </tr>
          </table>

          <h3 style="color:#4a4a8a;margin-top:0;">Результат</h3>
          <div style="background:#f9f9fc;padding:15px 18px;border-left:4px solid #4a4a8a;border-radius:4px;">
            %s
          </div>

          %s

        </td></tr>

        <tr><td style="background:#f5f5f5;padding:14px 30px;text-align:center;font-size:12px;color:#888;">
          Автоматическое письмо от Telegram-бота Альфии Харисовой
        </td></tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`,
		data.TestName,
		userName,
		data.User.ID,
		data.CompletedAt.Format("02.01.2006 в 15:04 MST"),
		newlineToBreak(data.ResultMain),
		detailBlock,
	)
}
