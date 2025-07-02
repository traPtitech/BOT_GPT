package handler

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/gpt"
	"github.com/traPtitech/BOT_GPT/internal/rag"
	"github.com/traPtitech/BOT_GPT/internal/repository"

	"github.com/sashabaranov/go-openai"
)

func messageReceived(messageText, messagePlainText, channelID string) {
	if isStaging(channelID) {
		_, err := bot.PostMessageWithErr(channelID, "ステージング機能が有効です。")
		if err != nil {
			fmt.Println(err)
		}

		if containsReset(messageText) {
			rag.ChatReset(channelID)

			return
		}

		if containsModelCommand(messageText) {
			handleModelCommand(messageText, channelID)

			return
		}

		imagesBase64 := bot.GetBase64ImagesFromMessage(messageText)
		rag.Chat(channelID, messagePlainText, imagesBase64)

		return
	}

	if containsReset(messageText) {
		gpt.ChatReset(channelID)

		return
	}

	if containsModelCommand(messageText) {
		handleModelCommand(messagePlainText, channelID)

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

/*
/modelコマンドの検出
*/
func containsModelCommand(input string) bool {
	re := regexp.MustCompile(`(^|\s)/model($|\s)`)

	return re.MatchString(input)
}

/*
モデル選択コマンドの処理
*/
func handleModelCommand(messageText, channelID string) {
	// /model show - 現在のモデルを表示
	if strings.Contains(messageText, "/model show") {
		currentModel, err := repository.GetModelForChannel(channelID)
		if err != nil {
			currentModel = "gpt-4o"
		}
		_, err = bot.PostMessageWithErr(channelID, fmt.Sprintf("現在のモデル: %s", currentModel))
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	// /model list - 利用可能なモデルを一覧表示
	if strings.Contains(messageText, "/model list") {
		availableModels := "利用可能なモデル:\n- gpt-4o(デフォルト)\n- gpt-4.1\n- o3\n- o4-mini"
		_, err := bot.PostMessageWithErr(channelID, availableModels)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	// /model set <model_name> - モデルを設定
	if strings.Contains(messageText, "/model set") {
		parts := strings.Split(messageText, " ")
		if len(parts) < 3 {
			_, err := bot.PostMessageWithErr(channelID, "使用方法: /model set <model_name>\n例: /model set o4-mini")
			if err != nil {
				fmt.Println(err)
			}

			return
		}

		modelName := parts[2]
		validModels := map[string]bool{
			openai.GPT4o:  true,
			openai.GPT4Dot1: true,
			openai.O3:     true,
			openai.O4Mini: true,
		}

		if !validModels[modelName] {
			_, err := bot.PostMessageWithErr(channelID, "無効なモデルです。/model listで利用可能なモデルを確認してください。")
			if err != nil {
				fmt.Println(err)
			}

			return
		}

		err := repository.SetModelForChannel(channelID, modelName)
		if err != nil {
			_, err = bot.PostMessageWithErr(channelID, "モデルの設定に失敗しました: "+err.Error())
			if err != nil {
				fmt.Println(err)
			}

			return
		}

		_, err = bot.PostMessageWithErr(channelID, fmt.Sprintf("モデルを %s に設定しました。", modelName))
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	// ヘルプメッセージ
	helpMessage := "モデル選択コマンド:\n- /model show: 現在のモデルを表示\n- /model list: 利用可能なモデルを一覧表示\n- /model set <model_name>: モデルを設定"
	_, err := bot.PostMessageWithErr(channelID, helpMessage)
	if err != nil {
		fmt.Println(err)
	}
}

// ステージングチャンネルであることを確認する
func isStaging(channelID string) bool {
	staginChannelID, ok := os.LookupEnv("STAGING_CHANNEL_ID")
	if !ok {

		return false
	}

	return channelID == staginChannelID
}
