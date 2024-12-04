package telegram

import (
	"fmt"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	welcomeMessage = `
Please enter the your address 
`
)

func TestBot(t *testing.T) {
	bot, err := tgbotapi.NewBotAPI("7819699755:AAFdg4Oa72JIO5xk4dg_zdcaOtP-T49cvaw")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		// 7. 处理按钮点击事件
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "no" {
				continue
			}

			fmt.Println(update.CallbackQuery.Data)
			// 回应用户选择
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "connected")
			bot.Send(msg)

			// 确认回调处理完成
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "connected")
			bot.Request(callback)
		}

		if update.Message == nil {
			continue
		}

		// 检查是否是 命令
		if !update.Message.IsCommand() {
			fmt.Println(update.Message.Text)
			continue
		}

		switch update.Message.Command() {
		case "start":
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
				bot.Send(msg)
				continue
			}
			text := strings.ReplaceAll(update.Message.Text, "/start ", "")
			// 4. 创建 Inline Keyboard 按钮
			yesButton := tgbotapi.NewInlineKeyboardButtonData("✅ Yes", text)
			noButton := tgbotapi.NewInlineKeyboardButtonData("❌ No", "no")

			// 5. 创建 Inline Keyboard Markup
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(yesButton, noButton),
			)

			// 6. 发送带有按钮的消息
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Your account confirms to be connected to the wallet %s ?", text))
			msg.ReplyMarkup = keyboard

			bot.Send(msg)
		default:

		}
	}
}
