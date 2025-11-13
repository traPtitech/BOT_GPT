package gpt

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/gpt/tooling"
	"github.com/traPtitech/BOT_GPT/internal/repository"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
)

type FinishReason int

const (
	stop FinishReason = iota
	length
	errorHappen
)

var (
	blobs                    = [...]string{":blob_bongo:", ":blob_crazy_happy:", ":blob_grin:", ":blob_hype:", ":blob_love:", ":blob_lurk:", ":blob_pyon:", ":blob_pyon_inverse:", ":blob_slide:", ":blob_snowball_1:", ":blob_snowball_2:", ":blob_speedy_roll:", ":blob_speedy_roll_inverse:", ":blob_thinking:", ":blob_thinking_fast:", ":blob_thinking_portal:", ":blob_thinking_upsidedown:", ":blob_thonkang:", ":blob_thumbs_up:", ":blobblewobble:", ":blobenjoy:", ":blobglitch:", ":blobbass:", ":blobjam:", ":blobkeyboard:", ":bloblamp:", ":blobmaracas:", ":blobmicrophone:", ":blobthinksmart:", ":blobwobwork:", ":conga_party_thinking_blob:", ":Hyperblob:", ":party_blob:", ":partyparrot_blob:", ":partyparrot_blob_cat:"}
	amazed                   = [...]string{":amazed_fuzzy:", ":amazed_amazed_fuzzy:", ":amazed_god_enel:", ":amazed_hamutaro:"}
	warnings                 = [...]string{":warning:", ":ikura-hamu_shooting_warning:"}
	apiKey                   string
	baseURL                  string
	DefaultSystemRoleMessage                  = "あなたは日本の学生サークルである東京科学大学デジタル創作同好会traPの部内SNS「traQ」のユーザーを、楽しませる娯楽用途や勉強するための学習用途として、作られた対話型AIです。身内しかいないSNSで、ユーザーに緩く接してください。そして、ユーザーの言う事に出来る限り従うようにしてください。特定の指示がなければ、数式は\\[は使わずに$$で括った上で、\n - \\begin{align}(やequation,eqnarray,split等)は\\[は使わずに$$で括った上で、\\begin{aligned}を使う\n - \\newlineは\\\\等を使う\n - \\mboxは\\textを使う\n - \\(は使わずに$を使う\nようにしてください。"
	ChannelMessages                           = make(map[string]Message)
	toolProvider             tooling.Provider = tooling.NewStaticProvider(tooling.DefaultSpecs())
)

func SetToolProvider(p tooling.Provider) {
	if p == nil {
		toolProvider = tooling.NewStaticProvider(tooling.DefaultSpecs())

		return
	}

	toolProvider = p
}

type Message = []responses.ResponseInputItemUnionParam

const SystemString = "FirstSystemMessageを変更しました。/gptsys showで確認できます。\nFirstSystemMessageとは、常に履歴の一番最初に入り、最初にgptに情報や状況を説明するのに使用する文字列です"

func InitGPT() {
	apiKey = getAPIKey()
	baseURL = getAPIBaseURL()

	channelIDs, err := repository.GetChannelIDs()
	if err != nil {
		log.Fatal(err)
	}
	for _, channelID := range channelIDs {
		messages, err := repository.LoadMessages(channelID)
		if err != nil {
			log.Printf("Failed to load messages for channel %s: %v", channelID, err)
			// Initialize empty slice on error
			ChannelMessages[channelID] = make(Message, 0)
		} else {
			// 保存されたメッセージをResponse API形式に変換
			convertedMessages := convertChatMessagesToResponseItems(messages)
			ChannelMessages[channelID] = convertedMessages
		}
	}
}

// convertChatMessagesToResponseItems converts v2 ChatCompletionMessageParamUnion to Response API format
func convertChatMessagesToResponseItems(messages []openai.ChatCompletionMessageParamUnion) []responses.ResponseInputItemUnionParam {
	var result []responses.ResponseInputItemUnionParam

	for _, msg := range messages {
		if userMsg := msg.OfUser; userMsg != nil {
			if textContent := userMsg.Content.OfString; textContent.Valid() {
				result = append(result, responses.ResponseInputItemParamOfMessage(textContent.Value, "user"))
			} else if len(userMsg.Content.OfArrayOfContentParts) > 0 {
				var contentParams responses.ResponseInputMessageContentListParam
				for _, part := range userMsg.Content.OfArrayOfContentParts {
					if textPart := part.OfText; textPart != nil {
						contentParams = append(contentParams, responses.ResponseInputContentParamOfInputText(textPart.Text))
					} else if imagePart := part.OfImageURL; imagePart != nil {
						contentParams = append(contentParams, responses.ResponseInputContentParamOfInputText("[image]"))
					}
				}
				result = append(result, responses.ResponseInputItemParamOfMessage(contentParams, "user"))
			}
		} else if assistantMsg := msg.OfAssistant; assistantMsg != nil {
			if textContent := assistantMsg.Content.OfString; textContent.Valid() {
				result = append(result, responses.ResponseInputItemParamOfMessage(textContent.Value, "assistant"))
			}
		} else if systemMsg := msg.OfSystem; systemMsg != nil {
			if textContent := systemMsg.Content.OfString; textContent.Valid() {
				result = append(result, responses.ResponseInputItemParamOfMessage(textContent.Value, "system"))
			}
		}
	}

	return result
}

func getAPIKey() string {
	key, exist := os.LookupEnv("OPENAI_PROXY_API_KEY")
	if !exist {
		log.Fatal("OPENAI_PROXY_API_KEY is not set")
	}

	return key
}

func getAPIBaseURL() string {
	url, exist := os.LookupEnv("OPENAI_API_BASE_URL")
	if !exist {
		log.Fatal("OPENAI_API_BASE_URL is not set")
	}

	return url
}

func getRandomBlob() string {
	return blobs[rand.Intn(len(blobs))]
}

func getRandomAmazed() string {
	return amazed[rand.Intn(len(amazed))]
}

func getRandomWarning() string {
	return warnings[rand.Intn(len(warnings))]
}

func OpenAIStream(messages Message, model string, do func(string)) (responseMessage string, finishReason FinishReason, err error) {
	c := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	ctx := context.Background()

	tools, toolErr := toolProvider.Tools(ctx)
	if toolErr != nil {
		return "", errorHappen, fmt.Errorf("resolve tools: %w", toolErr)
	}

	// Response APIで全メッセージ履歴を使用
	req := responses.ResponseNewParams{
		Model: openai.ChatModel(model),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam(messages),
		},
		Tools: tools,
	}
	stream := c.Responses.NewStreaming(ctx, req)
	if stream.Err() != nil {
		return "", errorHappen, stream.Err()
	}
	defer stream.Close()

	for stream.Next() {
		ev := stream.Current()
		if stream.Err() != nil {
			do(responseMessage + getRandomWarning() + ":blobglitch: Error: " + fmt.Sprint(stream.Err()))
			finishReason = errorHappen

			break
		}

		switch ev.Type {
		case "response.output_text.delta":
			textDelta := ev.AsResponseOutputTextDelta()
			responseMessage += textDelta.Delta
			do(responseMessage)
		case "response.completed":
			finishReason = stop
			do(responseMessage)
		case "response.failed":
			failedEvent := ev.AsResponseFailed()
			do(responseMessage + getRandomAmazed() + ":blobglitch: Failed: " + failedEvent.Response.Error.Message)
			finishReason = errorHappen
		case "response.incomplete":
			incompleteEvent := ev.AsResponseIncomplete()
			// トークン上限の場合 (max_output_tokensはトークン上限を示す)
			if incompleteEvent.Response.IncompleteDetails.Reason == "max_output_tokens" {
				do(responseMessage + "\n" + getRandomAmazed() + "トークン(履歴を含む文字数)が上限に達しました。/resetを実行してください。")
				finishReason = length
			} else {
				do(responseMessage + getRandomWarning() + ":blobglitch: Incomplete: " + incompleteEvent.Response.IncompleteDetails.Reason)
				finishReason = errorHappen
			}
		case "error":
			errorEvent := ev.AsError()
			do(responseMessage + getRandomWarning() + ":blobglitch: Error: " + errorEvent.Message)
			finishReason = errorHappen
		}
	}

	// If we reach here, the stream ended without setting finishReason
	if finishReason == 0 {
		finishReason = stop
	}

	return
}

func Chat(channelID, newMessageText string, imageBase64 []string) {
	_, exist := ChannelMessages[channelID]
	if !exist {
		ChannelMessages[channelID] = make(Message, 0)
	}
	addSystemMessageIfNotExist(channelID, DefaultSystemRoleMessage)

	// チャンネルのモデル設定を取得
	model, err := repository.GetModelForChannel(channelID)
	if err != nil {
		model = string(openai.ChatModelGPT5Mini) // デフォルト
	}

	if len(imageBase64) >= 1 {
		addImageAndTextAsUser(channelID, newMessageText, imageBase64)
	} else {
		addMessageAsUser(channelID, newMessageText)
	}

	time.Sleep(50 * time.Millisecond)
	postMessage, err := bot.PostMessageWithErr(channelID, getRandomBlob()+":loading:")
	if err != nil {
		bot.EditMessage(postMessage.Id, getRandomAmazed()+"Error: "+fmt.Sprint(err))
	}

	responseMessage, finishReason, err := OpenAIStream(ChannelMessages[channelID], model, func(responseMessage string) {
		bot.EditMessage(postMessage.Id, responseMessage)
	})
	if err != nil {
		bot.EditMessage(postMessage.Id, fmt.Sprintf("before ChatCompletionStream error: %v\n", err))
	}

	addMessageAsAssistant(channelID, responseMessage)

	if finishReason == length {
		bot.EditMessage(postMessage.Id, responseMessage+"\n"+getRandomAmazed()+"トークン(履歴を含む文字数)が上限に達しました。/resetを実行してください。")
	}
	if finishReason == stop {
		bot.EditMessage(postMessage.Id, responseMessage)
	}
}

func ChatChangeSystemMessage(channelID, message string) {
	DefaultSystemRoleMessage = message
	bot.PostMessage(channelID, SystemString)
}

func ChatShowSystemMessage(channelID string) {
	bot.PostMessage(channelID, DefaultSystemRoleMessage)
}

func ChatReset(channelID string) {
	msg := bot.PostMessage(channelID, ":blobnom::loading:")
	ChannelMessages[channelID] = make(Message, 0)
	err := bot.EditMessageWithErr(msg.Id, ":done:")
	if err != nil {
		bot.EditMessage(msg.Id, "Error: "+fmt.Sprint(err))
	}

	if err = repository.DeleteMessages(channelID); err != nil {
		fmt.Println(err)
	}

	// モデルをデフォルトにリセット
	defaultModel := string(openai.ChatModelGPT5Mini)
	if err = repository.SetModelForChannel(channelID, defaultModel); err != nil {
		fmt.Printf("Failed to reset model for channel %s: %v\n", channelID, err)
	}
}

func addMessageAsUser(channelID, message string) {
	userMessage := responses.ResponseInputItemParamOfMessage(message, "user")
	ChannelMessages[channelID] = append(ChannelMessages[channelID], userMessage)

	index := len(ChannelMessages[channelID]) - 1
	if err := repository.SaveMessage(channelID, index, openai.UserMessage(message)); err != nil {
		fmt.Printf("Failed to save user message: %v\n", err)
	}
}

func addImageAndTextAsUser(channelID, message string, imageDataBase64 []string) {
	// 現時点では画像もテキストとして扱う
	var content responses.ResponseInputMessageContentListParam
	content = append(content, responses.ResponseInputContentParamOfInputText(message))
	if len(imageDataBase64) > 0 {
		imageText := fmt.Sprintf("[%d images attached]", len(imageDataBase64))
		content = append(content, responses.ResponseInputContentParamOfInputText(imageText))
	}

	userMessage := responses.ResponseInputItemParamOfMessage(content, "user")
	ChannelMessages[channelID] = append(ChannelMessages[channelID], userMessage)

	index := len(ChannelMessages[channelID]) - 1
	// repository用に旧形式で保存
	var parts []openai.ChatCompletionContentPartUnionParam

	parts = append(parts, openai.TextContentPart(message))

	for _, b64 := range imageDataBase64 {
		imageURL := "data:image/jpeg;base64," + b64
		imagePart := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: imageURL,
		})
		parts = append(parts, imagePart)
	}
	if err := repository.SaveMessage(channelID, index, openai.UserMessage(parts)); err != nil {
		fmt.Printf("Failed to save image message: %v\n", err)
	}
}

func addMessageAsAssistant(channelID, message string) {
	assistantMessage := responses.ResponseInputItemParamOfMessage(message, "assistant")
	ChannelMessages[channelID] = append(ChannelMessages[channelID], assistantMessage)

	index := len(ChannelMessages[channelID]) - 1
	if err := repository.SaveMessage(channelID, index, openai.AssistantMessage(message)); err != nil {
		fmt.Printf("Failed to save assistant message: %v\n", err)
	}
}

func addSystemMessageIfNotExist(channelID, message string) {
	// システムメッセージが既に存在するかチェック
	for _, msg := range ChannelMessages[channelID] {
		if msg.GetRole() != nil && *msg.GetRole() == "system" {
			return
		}
	}

	// システムメッセージが存在しない場合のみ先頭に追加
	systemMessage := responses.ResponseInputItemParamOfMessage(message, "system")
	ChannelMessages[channelID] = append([]responses.ResponseInputItemUnionParam{systemMessage}, ChannelMessages[channelID]...)

	index := 0
	if err := repository.SaveMessage(channelID, index, openai.SystemMessage(message)); err != nil {
		fmt.Printf("Failed to save system message: %v\n", err)
	}
}

//func updateSystemRoleMessage(channelID, message string) {
//	addSystemMessageIfNotExist(channelID, message)
//	ChannelMessages[channelID][0] = Message{
//		Role:    "system",
//		Content: message,
//	}
//
//	index := 0
//	if err := repository.SaveMessage(channelID, index, ChannelMessages[channelID][0]); err != nil {
//		fmt.Println(err)
//	}
//}

//func ChatDebug(channelID string) {
//	returnString := "```\n"
//	for _, m := range Messages {
//		chatText := regexp.MustCompile("```").ReplaceAllString(m.Content, "")
//		if len(chatText) >= 40 {
//			returnString += m.Role + ": " + chatText[:40] + "...\n"
//		} else {
//			returnString += m.Role + ": " + chatText + "\n"
//		}
//	}
//	returnString += "```"
//	_, err := bot.PostMessageWithErr(channelID, returnString)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
