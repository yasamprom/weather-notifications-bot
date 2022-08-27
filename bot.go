package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"time"
)

type State interface {
}

type CityChooseState struct {
	State
}

type TimerChooseState struct {
	State
}

var (
	bot, err  = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	cities, _ = create_db()
	state     = make(map[int64]State)
)

func processQuery(update *tgbotapi.Update) {
	switch update.Message.Text {
	case "/start":
		kb := main_kb()
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
		msg.ReplyMarkup = kb
		bot.Send(msg)
	case "Погода":
		filter := bson.D{{"name", update.Message.Chat.ID}}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var result struct {
			Value string
		}
		err = cities.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			// Do something when no record was found
			fmt.Println("Couldn't access or find value")
		}
		txt := weather(result.Value)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Погода в городе "+result.Value+":\n"+txt)
		bot.Send(msg)
		return
	case "Выбрать город":
		state[update.Message.Chat.ID] = CityChooseState{}
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Напиши город"))
		return
	case "Установить таймер":
		state[update.Message.Chat.ID] = TimerChooseState{}
		return
	}
	if val, ok := state[update.Message.Chat.ID]; ok {
		switch val {
		case CityChooseState{}:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := cities.InsertOne(ctx,
				bson.D{
					{"name", update.Message.Chat.ID},
					{"value", update.Message.Text},
				},

			)
			if err == nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сохранили город: "+update.Message.Text))
			} else {
				fmt.Println(err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "-"))
			}
		case TimerChooseState{}:
			// set timer for ID
		}
	}
}

func main() {
	if err != nil {
		return
	}

	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		go processQuery(&update)
	}
}
