package handler

import (
	"fmt"
	"os"
	"regexp"

	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/gpt"
)

func messageReceived(messageText, messagePlainText, channelID string) {
	if isStaging(channelID) {
		_, err := bot.PostMessageWithErr(channelID, "ステージング機能が有効です。")
		if err != nil {
			fmt.Println(err)
		}

		return
	}

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

// ステージングチャンネルであることを確認する
func isStaging(channelID string) bool {
	staginChannelID, ok := os.LookupEnv("STAGING_CHANNEL_ID")
	if !ok {
		return false
	}

	return channelID == staginChannelID
}
