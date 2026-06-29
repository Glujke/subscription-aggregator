package main

import (
	"context"
	"log"
	"os"

	"subscription-aggregator/internal/app"
)

// @title           Subscription Aggregator API
// @version         1.0
// @description     REST-сервис агрегации данных об онлайн-подписках пользователей.
// @host            localhost:8080
// @BasePath        /
func main() {
	if err := app.Run(context.Background()); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
