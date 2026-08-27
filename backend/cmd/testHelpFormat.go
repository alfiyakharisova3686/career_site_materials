package main

import (
	"fmt"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SessionHelpFormat struct {
	ChatID       int64
	CallbackData []string
	CurrentStep  int
	Resume       int
	Hypotheses   int
	Consulting   int
	Barriers     int
	Dismissal    int
	Support      int
	Q0Code       string
	Q3Code       string
}

type QuestionnaireHelpFormat struct {
	Name            string
	Description     string
	Questions       []string
	Answers         [][]string
	ResultTexts     map[string]string
	SecondaryName   map[string]string
	SecondaryReason map[string]string
}

func newSessionHelpFormat(chatID int64) {
	mu.Lock()
	defer mu.Unlock()

	sessionsHelpFormat[chatID] = &SessionHelpFormat{
		ChatID: chatID,
	}
}

func getSessionHelpFormat(chatID int64) *SessionHelpFormat {
	mu.RLock()
	defer mu.RUnlock()
	return sessionsHelpFormat[chatID]
}

func sendDescriptionHelpFormat(bot *tgbotapi.BotAPI, chatID int64, messageID int) {

	btns := buildKeyboard([]Btn{
		{Text: "Начать тест", Data: "begin_test_helpFormat"},
		{Text: "Вернуться в меню", Data: "btn_back_to_menu_helpFormat"},
	})

	text := fmt.Sprintf("<b>%s</b>\n\n%s", helpFormatTest.Name, helpFormatTest.Description)
	renderScreen(bot, chatID, messageID, text, btns)
}

func sendNextQuestionHelpFormat(bot *tgbotapi.BotAPI, chatID int64, messageID int, session *SessionHelpFormat) {
	btns := buildKeyboard([]Btn{})
	for i, answer := range helpFormatTest.Answers[session.CurrentStep] {
		newBtn := tgbotapi.NewInlineKeyboardButtonData(answer, fmt.Sprintf("nav_%d_%d", session.CurrentStep, i))
		btns.InlineKeyboard = append(btns.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(newBtn))
	}

	exitBtn := tgbotapi.NewInlineKeyboardButtonData("Вернуться в меню", "btn_back_to_menu_helpFormat")
	btns.InlineKeyboard = append(btns.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(exitBtn))

	text := fmt.Sprintf("<b>Вопрос %d из %d</b>\n\n%s", session.CurrentStep+1, len(helpFormatTest.Questions), helpFormatTest.Questions[session.CurrentStep])

	renderScreen(bot, chatID, messageID, text, btns)
}

func beginTestHelpFormat(bot *tgbotapi.BotAPI, chatID int64, question int, answer int) {
	session := getSessionHelpFormat(chatID)

	switch question {
	case 0: // Что лучше всего описывает вашу ситуацию сейчас?
		switch answer {
		case 0: // Не понимаю, куда двигаться дальше
			session.Consulting += 2
			session.Hypotheses += 1
			session.Q0Code = "CONSULTING"
		case 1: // Рассматриваю несколько направлений и не могу выбрать
			session.Consulting += 3
			session.Q0Code = "CONSULTING"
		case 2: // Знаю, чего хочу, но не решаюсь действовать
			session.Barriers += 3
			session.Q0Code = "BARRIERS"
		case 3: // Цель понятна, но сложно двигаться самостоятельно
			session.Support += 3
			session.Q0Code = "SUPPORT"
		case 4: // После увольнения трудно понять, что делать дальше
			session.Dismissal += 4
			session.Q0Code = "DISMISSAL"
		case 5: // Хочу улучшить резюме и получать больше откликов
			session.Resume += 4
			session.Q0Code = "RESUME"
		}

	case 1: // Насколько ясно вы понимаете своё дальнейшее направление?
		switch answer {
		case 0: // Пока совсем не понимаю
			session.Consulting += 2
			session.Hypotheses += 1
		case 1: // Есть несколько вариантов
			session.Consulting += 3
		case 2: // Направление понятно, но цель ещё размыта
			session.Consulting += 1
			session.Support += 1
		case 3: // Есть конкретная цель
			session.Support += 2
			session.Barriers += 1
		case 4: // Сейчас мне сложно думать о следующем шаге
			session.Dismissal += 3
		}

	case 2: // Что сейчас мешает больше всего?
		switch answer {
		case 0: // Не хватает понимания своих сильных сторон
			session.Hypotheses += 2
			session.Consulting += 1
		case 1: // Боюсь ошибиться с выбором
			session.Consulting += 2
			session.Barriers += 1
		case 2: // Тревога, неуверенность или страх оценки
			session.Barriers += 3
		case 3: // Обесцениваю свой опыт и достижения
			session.Barriers += 3
		case 4: // Нет понятного плана и регулярных действий
			session.Support += 3
		case 5: // Тяжело переживаю увольнение или прошлый опыт
			session.Dismissal += 4
		case 6: // Резюме не показывает мою реальную ценность
			session.Resume += 4
		}

	case 3: // Какой результат вам нужен в первую очередь?
		switch answer {
		case 0: // Получить идеи подходящих сфер и ролей
			session.Hypotheses += 2
			session.Consulting += 1
			session.Q3Code = "HYPOTHESES"
		case 1: // Выбрать одно основное направление
			session.Consulting += 3
			session.Q3Code = "CONSULTING"
		case 2: // Разобраться с тем, что мешает сделать шаг
			session.Barriers += 4
			session.Q3Code = "BARRIERS"
		case 3: // Вернуть опору после увольнения
			session.Dismissal += 4
			session.Q3Code = "DISMISSAL"
		case 4: // Составить план и двигаться к конкретной цели
			session.Support += 4
			session.Q3Code = "SUPPORT"
		case 5: // Понять, как улучшить резюме
			session.Resume += 4
			session.Q3Code = "RESUME"
		}

	case 4: // Какой формат вам сейчас комфортнее?
		switch answer {
		case 0: // Получить письменный разбор без встречи
			session.Hypotheses += 4
		case 1: // Одна консультация по конкретному вопросу
			session.Resume += 1
			session.Barriers += 1
		case 2: // Несколько встреч для более глубокого разбора
			session.Consulting += 1
			session.Barriers += 1
			session.Dismissal += 1
		case 3: // Регулярное сопровождение и поддержка
			session.Support += 4
		case 4: // Пока не знаю
			// Баллы не начислять
		}

	case 5: // Что вы готовы делать после получения рекомендаций?
		switch answer {
		case 0: // Самостоятельно изучить предложенные варианты
			session.Hypotheses += 3
		case 1: // Обсуждать гипотезы и выбирать вместе со специалистом
			session.Consulting += 3
		case 2: // Работать с внутренними причинами, которые удерживают
			session.Barriers += 4
		case 3: // Выполнять конкретные действия между встречами
			session.Support += 3
		case 4: // Сначала восстановиться и прийти в более устойчивое состояние
			session.Dismissal += 4
		case 5: // Внести конкретные изменения в резюме и продолжить поиск
			session.Resume += 3
		}
	}
}

func resultHelpFormat(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	session := getSessionHelpFormat(chatID)
	var text1, text2, primaryKey string
	points := map[string]int{"SUPPORT": session.Support, "HYPOTHESES": session.Hypotheses, "BARRIERS": session.Barriers, "DISMISSAL": session.Dismissal, "RESUME": session.Resume, "CONSULTING": session.Consulting}
	maxPoint := session.Support
	maxKey := "SUPPORT"

	for key, val := range points {
		if val > maxPoint {
			maxPoint = val
			maxKey = key
		}
	}

	switch {
	case session.Dismissal >= 8 && (slices.Contains(session.CallbackData, "q_0_4") || slices.Contains(session.CallbackData, "q_2_5") || slices.Contains(session.CallbackData, "q_3_3") || slices.Contains(session.CallbackData, "q_5_4")):
		text1 = helpFormatTest.ResultTexts["DISMISSAL"]
		primaryKey = "DISMISSAL"

	case maxKey == "RESUME" && (slices.Contains(session.CallbackData, "q_1_0") || slices.Contains(session.CallbackData, "q_1_1")):
		text1 = helpFormatTest.ResultTexts["CONSULTING"]
		primaryKey = "CONSULTING"
		text2 = fmt.Sprintf("<b>Также можно рассмотреть:</b> %s\n\n%s",
			helpFormatTest.SecondaryName["RESUME"],
			helpFormatTest.SecondaryReason["RESUME"])

	case slices.Contains(session.CallbackData, "q_4_0") && (session.Hypotheses > session.Consulting-2):
		text1 = helpFormatTest.ResultTexts["HYPOTHESES"]
		primaryKey = "HYPOTHESES"

	case (slices.Contains(session.CallbackData, "q_1_0") || slices.Contains(session.CallbackData, "q_1_1")) && slices.Contains(session.CallbackData, "q_4_0") && (session.Hypotheses < session.Consulting):
		text1 = helpFormatTest.ResultTexts["CONSULTING"]
		primaryKey = "CONSULTING"

	case (slices.Contains(session.CallbackData, "q_1_2") || slices.Contains(session.CallbackData, "q_1_3")) && (session.Barriers > session.Support):
		text1 = helpFormatTest.ResultTexts["BARRIERS"]
		primaryKey = "BARRIERS"

	case (slices.Contains(session.CallbackData, "q_1_2") || slices.Contains(session.CallbackData, "q_1_3")) && (session.Barriers < session.Support):
		text1 = helpFormatTest.ResultTexts["SUPPORT"]
		primaryKey = "SUPPORT"

	default:
		equalCount := 0
		for _, value := range points {
			if value == maxPoint {
				equalCount += 1
			}
		}

		switch {
		case equalCount >= 3:
			text1 = helpFormatTest.ResultTexts["NEUTRAL"]
			primaryKey = "NEUTRAL"

		case equalCount == 2:
			if points["HYPOTHESES"] == maxPoint && points["CONSULTING"] == maxPoint {
				if slices.Contains(session.CallbackData, "q_4_0") {
					maxKey = "HYPOTHESES"
				} else {
					maxKey = "CONSULTING"
				}
			} else {
				for key, value := range points {
					if value == maxPoint && key != maxKey {
						if session.Q0Code == key {
							maxKey = key
						} else if session.Q0Code != maxKey && session.Q3Code == key {
							maxKey = key
						}
						break
					}
				}
			}
			text1 = helpFormatTest.ResultTexts[maxKey]

		default:
			text1 = helpFormatTest.ResultTexts[maxKey]
		}
		primaryKey = maxKey
	}

	if text2 == "" && primaryKey != "DISMISSAL" && primaryKey != "NEUTRAL" {
		allowedPairs := map[string][]string{
			"CONSULTING": {"RESUME", "BARRIERS"},
			"SUPPORT":    {"BARRIERS"},
		}

		if allowed, ok := allowedPairs[primaryKey]; ok {
			for _, code := range allowed {
				score := points[code]
				if points[primaryKey]-score <= 2 && score > 0 {
					text2 = fmt.Sprintf("<b>Также можно рассмотреть:</b> %s\n\n%s", helpFormatTest.SecondaryName[code], helpFormatTest.SecondaryReason[code])
					break
				}
			}
		}
	}

	btns := buildKeyboard([]Btn{{Text: "В главное меню", Data: "btn_back_to_menu"}})
	urlAccount := tgbotapi.NewInlineKeyboardButtonURL("Обсудить разбор", "https://t.me/Harisova_Alfiya")
	urlChannel := tgbotapi.NewInlineKeyboardButtonURL("Подробнее о форматах работы", "https://t.me/proforientacia_alfiya/61")
	btns.InlineKeyboard = append(btns.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(urlAccount))
	btns.InlineKeyboard = append(btns.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(urlChannel))

	if text2 != "" {
		renderScreen(bot, chatID, messageID, text1, buildKeyboard([]Btn{}))
		renderScreen(bot, chatID, 0, text2, btns)
	} else {
		renderScreen(bot, chatID, messageID, text1, btns)
	}

}
