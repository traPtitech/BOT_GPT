package gpt

import (
	"context"
	"errors"
	"fmt"
	"github.com/pikachu0310/BOT_GPT/internal/bot"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"

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
	DefaultSystemRoleMessage = "あなたはサークルである東京工業大学デジタル創作同好会traPの部内SNS「traQ」のユーザーを、楽しませる娯楽用途や勉強するための学習用途として、BOTの中に作られたOpenAIの最新モデルGPT4oを用いた対話型AIです。身内しかいないSNSで、ユーザーに緩く接してください。そして、ユーザーの言う事に出来る限り従うようにしてください。"
)

type Message = openai.ChatCompletionMessage

var Messages = make([]Message, 0)

const GptSystemString = "FirstSystemMessageを変更しました。/gptsys showで確認できます。\nFirstSystemMessageとは、常に履歴の一番最初に入り、最初にgptに情報や状況を説明するのに使用する文字列です"

func InitGPT() {
	apiKey = getAPIKey()
}

func getAPIKey() string {
	key, exist := os.LookupEnv("OPENAI_API_KEY")
	if !exist {
		log.Fatal("OPENAI_API_KEY is not set")
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
	c := openai.NewClient(apiKey)
	ctx := context.Background()

	model := openai.GPT4o

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		do(fmt.Sprintf("ChatCompletionStream error: %v\n", err))
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
			err = errors.New("stream closed")
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

		if time.Now().Sub(lastDoTime) >= deltaTime {
			lastDoTime = time.Now()
			do(getRandomBlobAndAmazed() + responseMessage + ":loading:")
		}
	}
	addMessageAsAssistant(responseMessage)
	return
}

func Chat(channelID, newMessage string, imageBase64 []string) {
	if len(imageBase64) >= 1 {
		addImageAndTextAsUser(newMessage, imageBase64)
	} else {
		addMessageAsUser(newMessage)
	}
	updateSystemRoleMessage(DefaultSystemRoleMessage)
	time.Sleep(50 * time.Millisecond)
	postMessage, err := bot.PostMessageWithErr(channelID, getRandomBlob()+":loading:")
	if err != nil {
		bot.EditMessage(postMessage.Id, getRandomAmazed()+"Error: "+fmt.Sprint(err))
	}
	responseMessage, finishReason, err := OpenAIStream(Messages, func(responseMessage string) {
		bot.EditMessage(postMessage.Id, responseMessage)
	})
	if err != nil {
		fmt.Println(err)
	}

	if finishReason == length {
		bot.EditMessage(postMessage.Id, responseMessage+"\n"+getRandomAmazed()+"トークン(履歴を含む文字数)が上限に達しました。/resetを実行してください。")
	}

	if finishReason == stop {
		bot.EditMessage(postMessage.Id, responseMessage)
	}
}

func ChatChangeSystemMessage(channelID, message string) {
	DefaultSystemRoleMessage = message
	bot.PostMessage(channelID, GptSystemString)
}

func ChatShowSystemMessage(channelID string) {
	bot.PostMessage(channelID, DefaultSystemRoleMessage)
}

func ChatReset(channelID string) {
	msg := bot.PostMessage(channelID, ":blobnom::loading:")
	Messages = make([]Message, 0)
	err := bot.EditMessageWithErr(msg.Id, ":done:")
	if err != nil {
		bot.EditMessage(msg.Id, "Error: "+fmt.Sprint(err))
	}
	return
}

func addMessageAsUser(message string) {
	Messages = append(Messages, Message{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})
}

func addImageAndTextAsUser(message string, imageDataBase64 []string) {
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

	messagePart := Message{
		Role:         "user",
		MultiContent: parts,
	}

	Messages = append(Messages, messagePart)
}

func addMessageAsAssistant(message string) {
	Messages = append(Messages, Message{
		Role:    openai.ChatMessageRoleAssistant,
		Content: message,
	})
}

func addMessageAsSystem(message string) {
	Messages = append(Messages, Message{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	})
}

func addSystemMessageIfNotExist(message string) {
	for _, m := range Messages {
		if m.Role == "system" {
			return
		}
	}
	Messages = append([]Message{{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	}}, Messages...)
}

func updateSystemRoleMessage(message string) {
	addSystemMessageIfNotExist(message)
	Messages[0] = Message{
		Role:    "system",
		Content: message,
	}
}

func ChatDebug(channelID string) {
	returnString := "```\n"
	for _, m := range Messages {
		chatText := regexp.MustCompile("```").ReplaceAllString(m.Content, "")
		if len(chatText) >= 40 {
			returnString += m.Role + ": " + chatText[:40] + "...\n"
		} else {
			returnString += m.Role + ": " + chatText + "\n"
		}
	}
	returnString += "```"
	_, err := bot.PostMessageWithErr(channelID, returnString)
	if err != nil {
		fmt.Println(err)
	}
}
