package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type Session struct {
	TestID        string
	ChatID        int64
	CareerContext string
	Points        []int
	CurrentStep   int
}

type Questionnaire struct {
	TestID          string
	Name            string
	Description     string
	ContextQuestion string
	ContextAnswers  []string
	ContextResult   []string
	Questions       []string
	ResultTitle     []string
	ResultText      []string
	PracticalSteps  []string
	Scenarios       []string
}

type Btn struct {
	Text string
	Data string
}

func buildKeyboard(buttons []Btn) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, button := range buttons {
		newButton := tgbotapi.NewInlineKeyboardButtonData(button.Text, button.Data)
		rowButton := tgbotapi.NewInlineKeyboardRow(newButton)
		rows = append(rows, rowButton)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func renderScreen(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string, markup tgbotapi.InlineKeyboardMarkup) {
	if messageID == 0 {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = markup

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}

		return
	}

	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = &markup

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка редактирования сообщения : %v", err)
	}
}

func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	text := `Здравствуйте! Я Альфия Харисова.

Здесь собраны короткие опросники о карьере и внутренних барьерах. Они помогают лучше понять свою ситуацию и увидеть, что может мешать двигаться дальше.

Выберите раздел:`

	buttons := buildKeyboard([]Btn{
		{Text: "Синдром самозванца в карьере", Data: "start_test_imposter"},
		{Text: "Подобрать формат помощи", Data: "start_test_helpFormat"},
		{Text: "Обо мне", Data: "btn_about"},
	})

	urlBtn := tgbotapi.NewInlineKeyboardButtonURL("Перейти в Telegram-канал", "https://t.me/proforientacia_alfiya")
	buttons.InlineKeyboard = append(buttons.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(urlBtn))

	renderScreen(bot, chatID, messageID, text, buttons)
}

func sendAboutText(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	text := `<b>Обо мне</b>

Я Альфия Харисова — карьерный психолог, ЕМТ-коуч и HR-эксперт.

Более 15 лет я работала внутри компаний, прошла путь от менеджера по персоналу до HR-директора. Поэтому я понимаю и логику работодателя, и внутренние сложности специалиста, который проходит карьерные изменения.`

	text2 := `<b>Как я работаю</b>

Я помогаю понять, куда двигаться дальше, увидеть, что удерживает, и выстроить реалистичный карьерный маршрут.

В работе соединяю психологическое образование, карьерное консультирование, знание рынка труда и подход, который помогает пересматривать выводы из прошлого опыта.

<b>Мой принцип: «Прошлое влияет, но не определяет».</b>`

	text3 := `<b>С чем ко мне можно обратиться</b>

— непонятно, куда двигаться дальше или какую сферу выбрать;
— тревога, сомнения или прошлый опыт мешают сделать шаг;
— после увольнения трудно вернуть опору;
— цель уже ясна, но нужна поддержка в движении;
— нужно посмотреть на резюме глазами работодателя;
— подросток выбирает профессию и образование.

Формат зависит от запроса: от письменного разбора и одной встречи до консультационного пакета и сопровождения.`

	btns := buildKeyboard([]Btn{
		{Text: "Вернуться в меню", Data: "btn_back_to_menu"},
		{Text: "Синдром самозванца в карьере", Data: "start_test_imposter"},
		{Text: "Подобрать формат помощи", Data: "start_test_helpFormat"},
		{Text: "helpFormat", Data: "start_test_helpFormat"},
	})

	urlHelpFormat := tgbotapi.NewInlineKeyboardButtonURL("Форматы работы", "https://t.me/proforientacia_alfiya/61")
	urlChannel := tgbotapi.NewInlineKeyboardButtonURL("Перейти в Telegram-канал", "https://t.me/proforientacia_alfiya")

	btns.InlineKeyboard = append(btns.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(urlHelpFormat), tgbotapi.NewInlineKeyboardRow(urlChannel))
	renderScreen(bot, chatID, messageID, text, buildKeyboard([]Btn{}))
	renderScreen(bot, chatID, 0, text2, buildKeyboard([]Btn{}))
	renderScreen(bot, chatID, 0, text3, btns)
}

func sendExitConfirmation(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	text := "<b>Вы уверены, что хотите прервать тест?</b>\nТекущий прогресс будет потерян."

	btns := buildKeyboard([]Btn{
		{Text: "Да", Data: "btn_back_to_menu"},
		{Text: "Нет", Data: "btn_cancel_exit"},
	})

	renderScreen(bot, chatID, messageID, text, btns)
}

func sendTestDescription(bot *tgbotapi.BotAPI, chatID int64, messageID int, testID string) {
	test := allTests[testID]

	btns := buildKeyboard([]Btn{
		{Text: "Начать тест", Data: "begin_test_" + testID},
		{Text: "Вернуться в меню", Data: "btn_back_to_menu"},
	})

	text := fmt.Sprintf("%s\n\n%s", test.Name, test.Description)
	renderScreen(bot, chatID, messageID, text, btns)
}

func newSession(testID string, chatID int64) {
	test, exists := allTests[testID]
	if !exists {
		log.Printf("Попытка создать сессию для несуществующего теста: %s", testID)
	}

	mu.Lock()
	defer mu.Unlock()

	sessions[chatID] = &Session{
		TestID:        testID,
		ChatID:        chatID,
		CurrentStep:   1,
		CareerContext: test.ContextQuestion,
		Points:        make([]int, 0, len(test.Questions)),
	}
}

func getSession(chatID int64) *Session {
	mu.RLock()
	defer mu.RUnlock()
	return sessions[chatID]
}

func sendContextQuestion(bot *tgbotapi.BotAPI, chatID int64, messageID int, session *Session) {
	test := allTests[session.TestID]

	btns := make([]Btn, 0, len(test.ContextAnswers))
	for i, answer := range test.ContextAnswers {
		btn := Btn{
			Text: answer, Data: fmt.Sprintf("ctx_%d", i),
		}
		btns = append(btns, btn)
	}
	btns = append(btns, Btn{
		Text: "Вернуться в меню", Data: "btn_back_to_menu",
	})
	keyboard := buildKeyboard(btns)

	text := fmt.Sprintf("%s", test.ContextQuestion)
	renderScreen(bot, chatID, messageID, text, keyboard)
}

func sendNextQuestion(bot *tgbotapi.BotAPI, chatID int64, messageID int, session *Session) {
	test := allTests[session.TestID]
	question := test.Questions[session.CurrentStep-1]

	btns := buildKeyboard([]Btn{
		{Text: "Почти никогда", Data: "ans_0"},
		{Text: "Редко", Data: "ans_1"},
		{Text: "Иногда", Data: "ans_2"},
		{Text: "Почти всегда", Data: "ans_3"},

		{Text: "Вернуться в меню", Data: "btn_ask_exit"},
	})

	text := fmt.Sprintf("<b>Вопрос %d из %d</b>\n\n%s", session.CurrentStep, len(test.Questions), question)

	renderScreen(bot, chatID, messageID, text, btns)
}

func maxSum(values []int) (maxIndexes []int, maxValue int) {
	maxValue = -1

	for _, val := range values {
		if val > maxValue {
			maxValue = val
		}
	}

	for i, val := range values {
		if val > 3 && val > maxValue-2 {
			maxIndexes = append(maxIndexes, i)
		}
	}
	return maxIndexes, maxValue
}

func sendResultTestImposter(bot *tgbotapi.BotAPI, chatID int64, messageID int, session *Session) {
	resultSum := 0
	for _, val := range session.Points {
		resultSum += val
	}

	firstPart := session.Points[0] + session.Points[1] + session.Points[2]
	secondPart := session.Points[3] + session.Points[4] + session.Points[5]
	thirdPart := session.Points[6] + session.Points[7] + session.Points[8]
	fourthPart := session.Points[9] + session.Points[10] + session.Points[11]

	text1 := "<b>Ваш результат: "
	text2 := "<b>Что проявляется сильнее всего</b>\n\n"

	maxIndexes, maxSumm := maxSum([]int{firstPart, secondPart, thirdPart, fourthPart})

	switch {
	case maxSumm < 4:
		text2 = fmt.Sprintf("<b>С чего начать</b>\n\n%s", allTests[session.TestID].PracticalSteps[(resultSum/9)-(resultSum/36)])

	case len(maxIndexes) >= 3:
		text2 += fmt.Sprintf("%s\n\n%s", allTests[session.TestID].Scenarios[4], allTests[session.TestID].PracticalSteps[(resultSum/9)-(resultSum/36)])

	case len(maxIndexes) == 2:
		text2 += fmt.Sprintf("%s\n\n%s\n\n%s", allTests[session.TestID].Scenarios[maxIndexes[0]],
			allTests[session.TestID].Scenarios[maxIndexes[1]], allTests[session.TestID].PracticalSteps[(resultSum/9)-(resultSum/36)])

	case len(maxIndexes) == 1:
		text2 += fmt.Sprintf("%s\n\n%s", allTests[session.TestID].Scenarios[maxIndexes[0]], allTests[session.TestID].PracticalSteps[(resultSum/9)-(resultSum/36)])
	}

	switch {
	case resultSum < 9:
		text1 += fmt.Sprintf("%s</b>\n%d из 36\n\n%s\n\n%s",
			allTests[session.TestID].ResultTitle[0], resultSum, allTests[session.TestID].ResultText[0], session.CareerContext)

	case 8 < resultSum && resultSum < 18:
		text1 += fmt.Sprintf("%s</b>\n%dиз 36\n\n%s\n\n%s",
			allTests[session.TestID].ResultTitle[1], resultSum, allTests[session.TestID].ResultText[1], session.CareerContext)

	case 17 < resultSum && resultSum < 27:
		text1 += fmt.Sprintf("%s</b>\n%dиз 36\n\n%s\n\n%s",
			allTests[session.TestID].ResultTitle[2], resultSum, allTests[session.TestID].ResultText[2], session.CareerContext)

	case 26 < resultSum:
		text1 += fmt.Sprintf("%s</b>\n%dиз 36\n\n%s\n\n%s",
			allTests[session.TestID].ResultTitle[3], resultSum, allTests[session.TestID].ResultText[3], session.CareerContext)

	}

	renderScreen(bot, chatID, messageID, text1, buildKeyboard([]Btn{}))

	btns := buildKeyboard([]Btn{})
	urlAccount := tgbotapi.NewInlineKeyboardButtonURL("Обсудить разбор", "https://t.me/Harisova_Alfiya")
	urlChannel := tgbotapi.NewInlineKeyboardButtonURL("Перейти в канал", "https://t.me/proforientacia_alfiya")
	btns.InlineKeyboard = append(btns.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(urlAccount))
	btns.InlineKeyboard = append(btns.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(urlChannel))
	renderScreen(bot, chatID, 0, text2, btns)

	sendMainMenu(bot, chatID, messageID)

	mu.Lock()
	delete(sessions, chatID)
	mu.Unlock()
	sendMainMenu(bot, chatID, 0)
}

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	callback := tgbotapi.NewCallback(query.ID, "")
	bot.Request(callback)
	chatID := query.Message.Chat.ID
	data := query.Data
	messageID := query.Message.MessageID

	switch {
	case data == "btn_back_to_menu":
		mu.Lock()
		delete(sessions, chatID)
		mu.Unlock()
		sendMainMenu(bot, chatID, messageID)

	case data == "btn_back_to_menu_helpFormat":
		mu.Lock()
		delete(sessionsHelpFormat, chatID)
		mu.Unlock()
		sendMainMenu(bot, chatID, messageID)

	case data == "btn_ask_exit":
		sendExitConfirmation(bot, chatID, messageID)

	case data == "btn_cancel_exit":
		session := getSession(chatID)
		if session == nil {
			sendMainMenu(bot, chatID, messageID)
			return
		}
		sendNextQuestion(bot, chatID, messageID, session)

	case data == "btn_about":
		sendAboutText(bot, chatID, messageID)

	case strings.HasPrefix(data, "start_test_helpFormat"):
		sendDescriptionHelpFormat(bot, chatID, messageID)

	case strings.HasPrefix(data, "start_test_"):
		testID := strings.TrimPrefix(data, "start_test_")
		sendTestDescription(bot, chatID, messageID, testID)

	case strings.HasPrefix(data, "begin_test_helpFormat"):
		newSessionHelpFormat(chatID)
		sendNextQuestionHelpFormat(bot, chatID, messageID, sessionsHelpFormat[chatID])

	case strings.HasPrefix(data, "nav_"):
		parts := strings.Split(data, "_")
		session := getSessionHelpFormat(chatID)
		if session == nil {
			sendDescriptionHelpFormat(bot, chatID, messageID)
			return
		}
		question, _ := strconv.Atoi(parts[1])
		answer, _ := strconv.Atoi(parts[2])
		session.CallbackData = append(session.CallbackData, fmt.Sprintf("q_%d_%d", question, answer))
		beginTestHelpFormat(bot, chatID, question, answer)

		session.CurrentStep += 1
		if session.CurrentStep >= len(helpFormatTest.Questions) {
			resultHelpFormat(bot, chatID, messageID)
			mu.Lock()
			delete(sessionsHelpFormat, chatID)
			mu.Unlock()
		} else {
			sendNextQuestionHelpFormat(bot, chatID, messageID, session)
		}

	case strings.HasPrefix(data, "begin_test_"):
		testID := strings.TrimPrefix(data, "begin_test_")
		newSession(testID, chatID)

		sendContextQuestion(bot, chatID, messageID, getSession(chatID))

	case strings.HasPrefix(data, "ctx_"):
		answer, err := strconv.Atoi(strings.TrimPrefix(data, "ctx_"))
		if err != nil {
			log.Printf("Ошибка неправильный ответ на контекстный вопрос, %v", err)
		}

		session := getSession(chatID)
		if session == nil {
			sendTestDescription(bot, chatID, messageID, "imposter")
			return
		}

		text := "Вы отметили, что сильнее всего это чувствуется " + allTests[session.TestID].ContextResult[answer]

		session.CareerContext = text

		sendNextQuestion(bot, chatID, messageID, session)

	case strings.HasPrefix(data, "ans_"):
		answer, err := strconv.Atoi(strings.TrimPrefix(data, "ans_"))
		if err != nil {
			log.Printf("Ошибка неправильный ответ на вопрос, %v", err)
		}

		session := getSession(chatID)
		if session == nil {
			sendTestDescription(bot, chatID, messageID, "imposter")
			return
		}
		session.Points = append(session.Points, answer)
		session.CurrentStep++

		test := allTests[session.TestID]
		if session.CurrentStep > len(test.Questions) {
			sendResultTestImposter(bot, chatID, messageID, session)
		} else {
			sendNextQuestion(bot, chatID, messageID, session)
		}
	}

}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, берутся системные переменные")
	}

	apiKey := os.Getenv("TELEGRAM_BOT_API")
	if apiKey == "" {
		log.Fatal("Не найден api ключ бота")
	}

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("Не задан WEBHOOK_URL")
	}
	webhookURL = strings.TrimRight(webhookURL, "/")

	port := os.Getenv("WEBHOOK_PORT")
	if port == "" {
		port = "8443"
	}

	bot, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Успешно авторизован на аккаунте @%s", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(webhookURL + "/bot" + bot.Token)
	if err != nil {
		log.Fatal("Ошибка создания вебхука: ", err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatal("Ошибка установки вебхука: ", err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal("Ошибка получения информации о вебхуке: ", err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Последняя ошибка вебхука: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/bot" + bot.Token)

	log.Printf("Запуск сервера на порту %s", port)
	go func() {
		if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
			log.Fatal("Ошибка запуска HTTP сервера: ", err)
		}
	}()

	for update := range updates {

		if update.Message != nil {

			if update.Message.IsCommand() {

				switch update.Message.Command() {
				case "start":
					sendMainMenu(bot, update.Message.Chat.ID, 0)
				}

			}
			continue
		}

		if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		}
	}
}
