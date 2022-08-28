package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main_kb() tgbotapi.ReplyKeyboardMarkup {
	btn1 := tgbotapi.NewKeyboardButton(ADD_CITY)
	btn2 := tgbotapi.NewKeyboardButton(SET_TIMER)
	btn3 := tgbotapi.NewKeyboardButton(GET_WEATHER)
	kb := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn1, btn2, btn3})
	return kb
}
