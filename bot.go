package main

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/procyon-projects/chrono"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"strings"
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

// todo: saving to users list
var (
	bot, err              = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	cities, timers, users = create_db()
	state                 = make(map[int64]State)
	times                 = make(map[int64]chrono.ScheduledTask)
)

func extractCity(chatId int64) (string, error) {
	filter := bson.D{{"_id", chatId}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var result struct {
		Value string
	}
	err = cities.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		bot.Send(tgbotapi.NewMessage(chatId, "Добавьте город"))
		return "", errors.New("id not found in db")
	}
	return result.Value, nil
}

func extractTimer(chatId int64) (int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var result struct {
		Value string
	}
	filter := bson.D{{"_id", chatId}}
	err = timers.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return 0, 0, errors.New("not found id in db")
	}
	hour, _ := strconv.Atoi(strings.Split(result.Value, ":")[0])
	min, _ := strconv.Atoi(strings.Split(result.Value, ":")[1])
	return hour, min, nil
}

func registerNotification(chatId int64) bool {
	hour, min, err := extractTimer(chatId)
	if err != nil {
		return false
	}
	taskScheduler := chrono.NewDefaultTaskScheduler()
	now := time.Now()
	task, _ := taskScheduler.Schedule(
		func(ctx context.Context) {
			for true {
				city, err := extractCity(chatId)
				if err != nil {
					continue
				}
				bot.Send(
					tgbotapi.NewMessage(
						chatId,
						"Погода в городе "+city+":\n"+weather(city),
					),
				)
				time.Sleep(24 * time.Hour)
			}
		},
		chrono.WithTime(time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())),
	)
	if val, ok := times[chatId]; ok {
		val.Cancel()
	}
	times[chatId] = task
	return true
}

func processQuery(update *tgbotapi.Update) {
	switch update.Message.Text {
	case "/start":
		kb := main_kb()
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
		msg.ReplyMarkup = kb
		bot.Send(msg)
	case GET_WEATHER:
		city, err := extractCity(update.Message.Chat.ID)
		if err == mongo.ErrNoDocuments {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Добавьте город"))
			return
		}
		txt := weather(city)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Погода в городе "+city+":\n"+txt)
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
						"Теперь уведомления будут приходить в: "+update.Message.Text+" по МСК",
					),
				)
				registerNotification(update.Message.Chat.ID)
			} else {
				panic(err)
			}
		}
	}
}

func initial() {
	cursor, err := timers.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	for _, result := range results {
		str := fmt.Sprintf("%v", result.Map()["_id"])
		n, _ := strconv.ParseInt(str, 10, 64)
		registerNotification(n)
	}
}

func main() {
	if err != nil {
		return
	}
	initial()
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
