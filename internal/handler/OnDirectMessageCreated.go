package handler

import (
	"github.com/pikachu0310/BOT_GPT/internal/bot"
	"github.com/pikachu0310/BOT_GPT/internal/gpt"
	"github.com/traPtitech/traq-ws-bot/payload"
	"log"
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

		if len(plainTextWithoutMention) >= 6 && plainTextWithoutMention[:6] == "/reset" {
			gpt.ChatReset(p.Message.ChannelID)

			return
		}

		imagesBase64 := bot.GetBase64ImagesFromMessage(p.Message.Text)

		_ = bot.PostMessage(p.Message.ChannelID, "DMではあんまり沢山使わないでね。定期的な`/reset`を忘れない事。")

		gpt.Chat(p.Message.ChannelID, plainTextWithoutMention, imagesBase64)
	}
}
