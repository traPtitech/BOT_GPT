package handler

import (
	"fmt"
	"github.com/pikachu0310/BOT_GPT/internal/bot"
	"github.com/pikachu0310/BOT_GPT/internal/gpt"
	"github.com/traPtitech/traq-ws-bot/payload"
	"log"
)

func (h *Handler) MessageReceived() func(p *payload.MessageCreated) {
	return func(p *payload.MessageCreated) {
		log.Println("=================================================")
		log.Printf("MessageReceived()")
		log.Printf("Message created by %s\n", p.Message.User.DisplayName)
		log.Println("Message:" + p.Message.Text)
		log.Printf("Payload:"+"%+v", p)

		if p.Message.User.Bot {
			return
		}

		plainTextWithoutMention := bot.RemoveFirstBotId(p.Message.PlainText)

		// if first 5 text = debug
		if len(plainTextWithoutMention) >= 5 && plainTextWithoutMention[:5] == "debug" {
			fmt.Printf("embed: %#v\n", p.Message.Embedded)
			return
		}

		imagesBase64 := bot.GetBase64ImagesFromMessage(p.Message.Text)

		gpt.Chat(p.Message.ChannelID, plainTextWithoutMention, imagesBase64)
	}
}
