package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/pikachu0310/BOT_GPT/internal/bot"
	"github.com/pikachu0310/BOT_GPT/internal/gpt"
	"github.com/pikachu0310/BOT_GPT/internal/handler"
	"github.com/pikachu0310/BOT_GPT/internal/migration"
	"github.com/pikachu0310/BOT_GPT/internal/pkg/config"
	"github.com/pikachu0310/BOT_GPT/internal/repository"
	"log"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf(".env is not exist: %v", err)
	}

	bot.InitBot()
	gpt.InitGPT()

	// connect to database
	db, err := sqlx.Connect("mysql", config.MySQL().FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// migrate tables
	if err := migration.MigrateTables(db.DB); err != nil {
		log.Fatal(err)
	}

	// setup repository
	repo := repository.New(db)
	repository.InitDB(db)

	// setup handler
	h := handler.New(repo)

	// setup bot
	traQBot := bot.GetBot()
	traQBot.OnMessageCreated(h.MessageReceived())
	traQBot.OnDirectMessageCreated(h.DirectMessageReceived())

	if err := traQBot.Start(); err != nil {
		panic(err)
	}
}
