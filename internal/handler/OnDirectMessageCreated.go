package handler

import (
	"log"

	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/pkg/formatter"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func (h *Handler) DirectMessageReceived() func(p *payload.DirectMessageCreated) {
	return func(p *payload.DirectMessageCreated) {
		log.Println("=================================================")
		log.Printf("DirectMessageReceived()")
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

		if p.Message.User.Name != "pikachu" {
			_ = bot.PostMessage(p.Message.ChannelID, "DMではあんまり沢山使わないでね。定期的な`/reset`を忘れない事。")
		}

		messageReceived(p.Message.Text, formattedMessage, p.Message.ChannelID)
	}
}
