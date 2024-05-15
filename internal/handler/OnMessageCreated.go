package handler

import (
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

		gpt.Chat(p.Message.ChannelID, p.Message.PlainText)
	}
}
