package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	bot, err              = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	cities, timers, users = create_db()
	state                 = make(map[int64]State)
)

func processQuery(update *tgbotapi.Update) {
	switch update.Message.Text {
	case "/start":
		kb := main_kb()
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
		msg.ReplyMarkup = kb
		bot.Send(msg)
	case GET_WEATHER:
		filter := bson.D{{"_id", update.Message.Chat.ID}}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var result struct {
			Value string
		}
		err = cities.FindOne(ctx, filter).Decode(&result)
		if err == mongo.ErrNoDocuments {
			// Do something when no record was found
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Добавьте город"))
			return
		}
		txt := weather(result.Value)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Погода в городе "+result.Value+":\n"+txt)
		bot.Send(msg)
		return
	case ADD_CITY:
		state[update.Message.Chat.ID] = CityChooseState{}
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Напиши город"))
		return
	case SET_TIMER:
		state[update.Message.Chat.ID] = TimerChooseState{}
		bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Пришли время, в которое ты хочешь получать уведомление в формате HH:MM"),
		)
		return
	}
	if val, ok := state[update.Message.Chat.ID]; ok {
		switch val {
		case CityChooseState{}:
			delete(state, update.Message.Chat.ID)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			filter := bson.D{
				{"_id", update.Message.Chat.ID},
			}
			upd := bson.D{
				{"_id", update.Message.Chat.ID},
				{"value", update.Message.Text},
			}
			var updatedDocument bson.M
			err = cities.FindOneAndReplace(ctx, filter, upd).Decode(&updatedDocument)
			if err == mongo.ErrNoDocuments {
				_, err = cities.InsertOne(ctx, upd)
			}
			if err == nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сохранили город: "+update.Message.Text))
			} else {
				panic(err)
			}
		case TimerChooseState{}:
			delete(state, update.Message.Chat.ID)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			filter := bson.D{
				{"_id", update.Message.Chat.ID},
			}
			upd := bson.D{
				{"_id", update.Message.Chat.ID},
				{"value", update.Message.Text},
			}
			var updatedDocument bson.M
			err = timers.FindOneAndReplace(ctx, filter, upd).Decode(&updatedDocument)
			if err == mongo.ErrNoDocuments {
				_, err = timers.InsertOne(ctx, upd)
			}
			if err == nil {
				bot.Send(
					tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Уведомления будут приходить в: "+update.Message.Text,
					),
				)
			} else {
				panic(err)
			}
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
