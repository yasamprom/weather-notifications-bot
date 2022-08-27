package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main_kb() tgbotapi.ReplyKeyboardMarkup {
	btn1 := tgbotapi.NewKeyboardButton("Выбрать город")
	btn2 := tgbotapi.NewKeyboardButton("Установить таймер")
	btn3 := tgbotapi.NewKeyboardButton("Погода")
	kb := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn1, btn2, btn3})
	return kb
}
