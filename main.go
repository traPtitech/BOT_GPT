package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/gpt"
	"github.com/traPtitech/BOT_GPT/internal/handler"
	"github.com/traPtitech/BOT_GPT/internal/migration"
	"github.com/traPtitech/BOT_GPT/internal/pkg/config"
	"github.com/traPtitech/BOT_GPT/internal/rag"
	"github.com/traPtitech/BOT_GPT/internal/repository"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf(".env is not exist: %v", err)
	}

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

	// initialize bot and gpt
	bot.InitBot()
	gpt.InitGPT()
	rag.InitGPT()

	// setup bot
	traQBot := bot.GetBot()
	traQBot.OnMessageCreated(h.MessageReceived())
	traQBot.OnDirectMessageCreated(h.DirectMessageReceived())

	if err := traQBot.Start(); err != nil {
		panic(err)
	}
}
