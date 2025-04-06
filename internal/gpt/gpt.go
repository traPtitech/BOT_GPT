package gpt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/BOT_GPT/internal/repository"

	"github.com/sashabaranov/go-openai"
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
	blobsAndAmazed           = append(blobs[:], amazed[:]...)
	warnings                 = [...]string{":warning:", ":ikura-hamu_shooting_warning:"}
	apiKey                   string
	DefaultSystemRoleMessage = "あなたはサークルである東京工業大学デジタル創作同好会traPの部内SNS「traQ」のユーザーを、楽しませる娯楽用途や勉強するための学習用途として、BOTの中に作られたOpenAIの最新モデルGPT4oを用いた対話型AIです。身内しかいないSNSで、ユーザーに緩く接してください。そして、ユーザーの言う事に出来る限り従うようにしてください。特定の指示がなければ、数式は\\[は使わずに$$で括った上で、\n - \\begin{align}(やequation,eqnarray,split等)は\\[は使わずに$$で括った上で、\\begin{aligned}を使う\n - \\newlineは\\\\等を使う\n - \\mboxは\\textを使う\n - \\(は使わずに$を使う\nようにしてください。"
	ChannelMessages          = make(map[string][]Message)
)

type Message = openai.ChatCompletionMessage

const SystemString = "FirstSystemMessageを変更しました。/gptsys showで確認できます。\nFirstSystemMessageとは、常に履歴の一番最初に入り、最初にgptに情報や状況を説明するのに使用する文字列です"

func InitGPT() {
	apiKey = getAPIKey()

	channelIDs, err := repository.GetChannelIDs()
	if err != nil {
		log.Fatal(err)
	}
	for _, channelID := range channelIDs {
		messages, err := repository.LoadMessages(channelID)
		if err != nil {
			log.Fatal(err)
		}
		ChannelMessages[channelID] = messages
	}
}

func getAPIKey() string {
	key, exist := os.LookupEnv("OPENAI_PROXY_API_KEY")
	if !exist {
		log.Fatal("OPENAI_PROXY_API_KEY is not set")
	}

	return key
}

func getRandomBlob() string {
	return blobs[rand.Intn(len(blobs))]
}

func getRandomAmazed() string {
	return amazed[rand.Intn(len(amazed))]
}

func getRandomBlobAndAmazed() string {
	return blobsAndAmazed[rand.Intn(len(blobsAndAmazed))]
}

func getRandomWarning() string {
	return warnings[rand.Intn(len(warnings))]
}

func OpenAIStream(messages []Message, do func(string)) (responseMessage string, finishReason FinishReason, err error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://llm-proxy.trap.jp"
	c := openai.NewClientWithConfig(config)
	ctx := context.Background()

	model := openai.GPT4o

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return
	}
	defer stream.Close()

	deltaTime := 500 * time.Millisecond
	lastDoTime := time.Now()
	for {
		response, err := stream.Recv()

		if err != nil {
			do(responseMessage + getRandomWarning() + ":blobglitch: Error: " + fmt.Sprint(err))
			finishReason = errorHappen

			break
		}
		if errors.Is(err, io.EOF) {
			_ = errors.New("stream closed")
			fmt.Println("steam closed")
			finishReason = errorHappen

			break
		}

		if response.Choices[0].FinishReason == "stop" {
			time.Sleep(200 * time.Millisecond)
			do(responseMessage)
			finishReason = stop

			break
		} else if response.Choices[0].FinishReason == "length" {
			do(responseMessage + "\n" + getRandomAmazed() + "トークン(履歴を含む文字数)が上限に達しました。/resetを実行してください。")
			finishReason = length

			break
		}

		responseMessage += response.Choices[0].Delta.Content

		if time.Since(lastDoTime) >= deltaTime {
			lastDoTime = time.Now()
			do(getRandomBlobAndAmazed() + responseMessage + ":loading:")
		}
	}

	return
}

func Chat(channelID, newMessageText string, imageBase64 []string) {
	_, exist := ChannelMessages[channelID]
	if !exist {
		ChannelMessages[channelID] = make([]Message, 0)
	}
	addSystemMessageIfNotExist(channelID, DefaultSystemRoleMessage)

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

	responseMessage, finishReason, err := OpenAIStream(ChannelMessages[channelID], func(responseMessage string) {
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
	ChannelMessages[channelID] = make([]Message, 0)
	err := bot.EditMessageWithErr(msg.Id, ":done:")
	if err != nil {
		bot.EditMessage(msg.Id, "Error: "+fmt.Sprint(err))
	}

	if err = repository.DeleteMessages(channelID); err != nil {
		fmt.Println(err)
	}
}

func addMessageAsUser(channelID, message string) {
	newMessage := Message{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}
	ChannelMessages[channelID] = append(ChannelMessages[channelID], newMessage)

	index := len(ChannelMessages[channelID]) - 1
	if err := repository.SaveMessage(channelID, index, newMessage); err != nil {
		fmt.Println(err)
	}
}

func addImageAndTextAsUser(channelID, message string, imageDataBase64 []string) {
	var parts []openai.ChatMessagePart

	parts = append(parts, openai.ChatMessagePart{
		Type: openai.ChatMessagePartTypeText,
		Text: message,
	})

	for _, b64 := range imageDataBase64 {
		imageURL := "data:image/jpeg;base64," + b64
		imagePart := openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: imageURL,
			},
		}
		parts = append(parts, imagePart)
	}

	newMessage := Message{
		Role:         "user",
		MultiContent: parts,
	}

	ChannelMessages[channelID] = append(ChannelMessages[channelID], newMessage)

	index := len(ChannelMessages[channelID]) - 1
	if err := repository.SaveMessage(channelID, index, newMessage); err != nil {
		fmt.Println(err)
	}
}

func addMessageAsAssistant(channelID, message string) {
	newMessage := Message{
		Role:    openai.ChatMessageRoleAssistant,
		Content: message,
	}
	ChannelMessages[channelID] = append(ChannelMessages[channelID], newMessage)

	index := len(ChannelMessages[channelID]) - 1
	if err := repository.SaveMessage(channelID, index, newMessage); err != nil {
		fmt.Println(err)
	}
}

func addSystemMessageIfNotExist(channelID, message string) {
	for _, m := range ChannelMessages[channelID] {
		if m.Role == "system" {
			return
		}
	}
	ChannelMessages[channelID] = append([]Message{{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	}}, ChannelMessages[channelID]...)

	index := 0
	if err := repository.SaveMessage(channelID, index, ChannelMessages[channelID][0]); err != nil {
		fmt.Println(err)
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
