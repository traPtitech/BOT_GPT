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

		textEmbedFormatted := formatter.FormatEmbeds(p.Message.Text)
		textWithoutMention := bot.RemoveFirstBotID(textEmbedFormatted)
		formattedMessage, err := formatter.FormatQuotedMessage(p.Message.User.ID, textWithoutMention)
		if err != nil {
			log.Printf("Error formatting quoted message: %v\n", err)
			formattedMessage = textWithoutMention
		}

		messageReceived(p.Message.Text, formattedMessage, p.Message.ChannelID)
	}
}
