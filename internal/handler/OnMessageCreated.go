package handler

import (
	"log"

	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/pkg/formatter"
	"github.com/traPtitech/traq-ws-bot/payload"
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
		formattedMessage, err := formatter.FormatQuotedMessage(p.Message.User.ID, plainTextWithoutMention)
		if err != nil {
			log.Printf("Error formatting quoted message: %v\n", err)
			formattedMessage = plainTextWithoutMention
		}

		messageReceived(p.Message.Text, formattedMessage, p.Message.ChannelID)
	}
}
