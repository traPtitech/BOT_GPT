package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

var (
	Bot  *traqwsbot.Bot
	Info *traq.MyUserDetail
)

func InitBot() {
	token := GetToken()

	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: token,
	})
	if err != nil {
		log.Fatalf("error: Bot変数が作れなかった!: %v", err)
	}

	botInfo, res, err := bot.API().MeApi.GetMe(context.Background()).Execute()
	if err != nil || res.StatusCode != 200 {
		log.Fatalf("error: 自分の情報を取得できませんでした: %v", err)
	}

	Bot = bot
	Info = botInfo
}

func RemoveFirstBotID(input string) string {
	// Bot IDでの検索を試行
	BotID := Info.Id
	mentionPattern := "@" + BotID
	index := strings.Index(input, mentionPattern)

	// Bot IDが見つからない場合は、Bot名での検索を試行
	if index == -1 {
		botName := Info.Name
		mentionPattern = "@" + botName
		index = strings.Index(input, mentionPattern)
	}

	if index == -1 {
		return input
	}

	// メンション部分を削除し、余分な空白を整理
	result := input[:index] + input[index+len(mentionPattern):]
	result = strings.TrimSpace(result)

	return result
}

func GetToken() (token string) {
	token, exist := os.LookupEnv("BOT_TOKEN")
	if !exist {
		log.Fatal("error: BOT_TOKENが設定されていません")
	}

	return token
}

func GetBot() *traqwsbot.Bot {
	return Bot
}

func Join(ChannelID string) error {
	bot := GetBot()
	_, err := bot.API().BotApi.LetBotJoinChannel(context.Background(), Info.Id).PostBotActionJoinRequest(traq.PostBotActionJoinRequest{ChannelId: ChannelID}).Execute()

	return err
}

func Leave(ChannelID string) error {
	bot := GetBot()
	_, err := bot.API().BotApi.LetBotLeaveChannel(context.Background(), Info.Id).PostBotActionLeaveRequest(traq.PostBotActionLeaveRequest{ChannelId: ChannelID}).Execute()

	return err
}

func IsBotJoined(ChannelID string) (bool, error) {
	bot := GetBot()
	bots, _, err := bot.API().BotApi.GetChannelBots(context.Background(), ChannelID).Execute()
	if err != nil {
		return false, err
	}
	for _, bot := range bots {
		if bot.Id == Info.Id {
			return true, nil
		}
	}

	return false, nil
}

func GetBots() []traq.Bot {
	bot := GetBot()
	Bots, _, err := bot.API().BotApi.GetBots(context.Background()).Execute()
	if err != nil {
		fmt.Println(err)
	}

	return Bots
}

//func botToUser(bot traq.Bot) traq.User {
//	user := traq.User{
//		Id:          bot.Id,
//		Name:        bot.BotUserId,
//		DisplayName: "",
//		IconFileId:  "",
//		Bot:         true,
//		State:       botStateToUserState(bot.State),
//		UpdatedAt:   bot.UpdatedAt,
//	}
//
//	return user
//}
//
//func botStateToUserState(botState traq.BotState) traq.UserAccountState {
//	switch botState {
//	case traq.BOTSTATE_deactivated:
//		return traq.USERACCOUNTSTATE_deactivated
//	case traq.BOTSTATE_active:
//		return traq.USERACCOUNTSTATE_active
//	}
//
//	return traq.USERACCOUNTSTATE_suspended
//}
