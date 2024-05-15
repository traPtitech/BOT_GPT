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
	blobs          = [...]string{":blob_bongo:", ":blob_crazy_happy:", ":blob_grin:", ":blob_hype:", ":blob_love:", ":blob_lurk:", ":blob_pyon:", ":blob_pyon_inverse:", ":blob_slide:", ":blob_snowball_1:", ":blob_snowball_2:", ":blob_speedy_roll:", ":blob_speedy_roll_inverse:", ":blob_thinking:", ":blob_thinking_fast:", ":blob_thinking_portal:", ":blob_thinking_upsidedown:", ":blob_thonkang:", ":blob_thumbs_up:", ":blobblewobble:", ":blobenjoy:", ":blobglitch:", ":blobbass:", ":blobjam:", ":blobkeyboard:", ":bloblamp:", ":blobmaracas:", ":blobmicrophone:", ":blobthinksmart:", ":blobwobwork:", ":conga_party_thinking_blob:", ":Hyperblob:", ":party_blob:", ":partyparrot_blob:", ":partyparrot_blob_cat:"}
	amazed         = [...]string{":amazed_fuzzy:", ":amazed_amazed_fuzzy:", ":amazed_god_enel:", ":amazed_hamutaro:"}
	blobsAndAmazed = append(blobs[:], amazed[:]...)
	warnings       = [...]string{":warning:", ":ikura-hamu_shooting_warning:"}
	apiKey         string
)

type Message = openai.ChatCompletionMessage

var Messages = make([]Message, 0)
var SystemRoleMessage = "あなたはサークルである東京工業大学デジタル創作同好会traPの部内SNS、traQのユーザーを楽しませる娯楽用途のBOTの中に作られた、openaiのモデルgpt-3.5-turboを用いた対話型AIです。身内しかいないSNSで、ユーザーに緩く接してください。そして、ユーザーの言う事に出来る限り従うようにしてください。"

const GptSystemString = "FirstSystemMessageを変更しました。/gptsys showで確認できます。\nFirstSystemMessageとは、常に履歴の一番最初に入り、最初にgptに情報や状況を説明するのに使用する文字列です"

func InitGPT() {
	apiKey = getApiKey()
}

func getApiKey() string {
	key, exist := os.LookupEnv("OPENAI_API_KEY")
	if !exist {
		log.Fatal("OPENAI_API_KEY is not set")
	}
	return key
}

func OpenAIStream(messages []Message, do func(string)) (responseMessage string, finishReason FinishReason, err error) {
	c := openai.NewClient(apiKey)
	ctx := context.Background()

	var model string
	model = openai.GPT4o

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	deltaTime := 500 * time.Millisecond
	lastDoTime := time.Now()
	for {
		response, err := stream.Recv()

		if err != nil {
			do(responseMessage + warnings[rand.Intn(len(warnings))] + ":blobglitch: Error: " + fmt.Sprint(err))
			finishReason = errorHappen
			break
		}
		if errors.Is(err, io.EOF) {
			err = errors.New("stream closed")
			finishReason = errorHappen
			break
		}

		if response.Choices[0].FinishReason == "stop" {
			time.Sleep(200 * time.Millisecond)
			do(responseMessage)
			finishReason = stop
			break
		} else if response.Choices[0].FinishReason == "length" {
			do(responseMessage + "\n" + amazed[rand.Intn(len(amazed))] + "トークン(履歴を含む文字数)が上限に達したので履歴の最初のメッセージを削除して続きを出力します:loading:")
			finishReason = length
			break
		}

		responseMessage += response.Choices[0].Delta.Content

		if time.Now().Sub(lastDoTime) >= deltaTime {
			lastDoTime = time.Now()
			do(blobsAndAmazed[rand.Intn(len(blobs))] + responseMessage + ":loading:")
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
	updateSystemRoleMessage(SystemRoleMessage)
	postMessage, err := bot.PostMessageWithErr(channelID, blobs[rand.Intn(len(blobs))]+":loading:")
	if err != nil {
		fmt.Println(err)
	}
	responsedMessage, finishReason, err := OpenAIStream(Messages, func(responseMessage string) {
		bot.EditMessage(postMessage.Id, responseMessage)
	})
	if err != nil {
		fmt.Println(err)
	}

	// finishReasonがlength以外になるまで、最大5回まで繰り返す
	for i := 0; i < 5 && finishReason == length; i++ {
		time.Sleep(500 * time.Millisecond)
		nowPostMessage := responsedMessage
		if len(Messages) >= 5 {
			Messages = Messages[3:]
		} else if len(Messages) >= 3 {
			Messages = Messages[2:]
		} else if len(Messages) >= 1 {
			Messages = Messages[1:]
		}
		addMessageAsUser("先ほどのあなたのメッセージが途中で途切れてしまっているので、続きだけを出力してください。")
		responsedMessage, finishReason, err = OpenAIStream(Messages, func(responseMessage string) {
			bot.EditMessage(postMessage.Id, nowPostMessage+"\n"+responseMessage)
		})
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func ChatChangeSystemMessage(channelID, message string) {
	SystemRoleMessage = message
	bot.PostMessage(channelID, GptSystemString)
}

func ChatShowSystemMessage(channelID string) {
	bot.PostMessage(channelID, SystemRoleMessage)
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
