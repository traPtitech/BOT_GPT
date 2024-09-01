package handler

import (
	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/gpt"
	"regexp"
)

func messageReceived(messageText, messagePlainText, channelID string) {
	if containsReset(messageText) {
		gpt.ChatReset(channelID)

		return
	}

	imagesBase64 := bot.GetBase64ImagesFromMessage(messageText)
	gpt.Chat(channelID, messagePlainText, imagesBase64)
}

/*
/resetの前に空白がある、または文字列の最初であること。
/resetの後に空白がある、または文字列の最後であること。
*/
func containsReset(input string) bool {
	re := regexp.MustCompile(`(^|\s)/reset($|\s)`)

	return re.MatchString(input)
}
