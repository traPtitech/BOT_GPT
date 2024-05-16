package handler

import (
	"github.com/pikachu0310/BOT_GPT/internal/bot"
	"github.com/traPtitech/traq-ws-bot/payload"
	"log"
)

func (h *Handler) MessageReceived() func(p *payload.MessageCreated) {
	return func(p *payload.MessageCreated) {
		log.Println("=================================================")
		log.Printf("MessageReceived()")
		log.Printf("Payload:"+"%+v", p)

		if p.Message.User.Bot {
			return
		}

		plainTextWithoutMention := bot.RemoveFirstBotID(p.Message.PlainText)

		messageReceived(p.Message.Text, plainTextWithoutMention, p.Message.ChannelID)
	}
}
