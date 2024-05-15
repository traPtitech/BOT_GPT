package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/pikachu0310/BOT_GPT/internal/bot"
	"github.com/pikachu0310/BOT_GPT/internal/handler"
	"github.com/pikachu0310/BOT_GPT/internal/pkg/config"
	"github.com/pikachu0310/BOT_GPT/internal/repository"
	"log"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("error: .env is not exist: %v", err)
	}

	// connect to database
	db, err := sqlx.Connect("mysql", config.MySQL().FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// setup repository
	repo := repository.New(db)

	// setup routes
	h := handler.New(repo)

	// setup bot
	traQBot := bot.GetBot()
	traQBot.OnMessageCreated(h.MessageReceived())
}
