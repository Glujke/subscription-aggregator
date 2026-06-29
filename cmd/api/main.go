package main

import (
	"context"
	"log"
	"os"

	"subscription-aggregator/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
