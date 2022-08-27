package main

import (
	"fmt"
	owm "github.com/briandowns/openweathermap"
	"log"
	"os"
)

func weather(city string) string {
	w, err := owm.NewCurrent("K", "RU", os.Getenv("API_WEATHER")) // 4dc0a1599c0c36727c2a66b5692ad590
	if err != nil {
		log.Fatalln(err)
	}
	w.CurrentByName(city)
	fmt.Println(w)
	fmt.Println(w.Key)
	temp := fmt.Sprintf("%.2f", w.Main.Temp-270)
	prob := fmt.Sprintf("%.2f", w.Rain.ThreeH)
	clouds := fmt.Sprintf("%d", w.Clouds.All)
	return "Температура: " + temp + "\nОсадки в мм: " + prob + "\nОблачность в %: " + clouds
}
