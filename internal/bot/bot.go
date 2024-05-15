package bot

import (
	"context"
	"fmt"
	"github.com/traPtitech/go-traq"
	_ "github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"log"
	"os"
)

var (
	Bot  *traqwsbot.Bot
	Info *traq.MyUserDetail
)

func init() {
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
	Info = botInfo
}

func GetToken() (token string) {
	token, exist := os.LookupEnv("BOT_TOKEN")
	if !exist {
		log.Fatal("error: BOT_TOKENが設定されていません")
	}
	return token
}

func GetBot() (bot *traqwsbot.Bot) {
	return bot
}

func BotJoin(ChannelID string) error {
	bot := GetBot()
	_, err := bot.API().BotApi.LetBotJoinChannel(context.Background(), Info.Id).PostBotActionJoinRequest(traq.PostBotActionJoinRequest{ChannelId: ChannelID}).Execute()
	return err
}

func BotLeave(ChannelID string) error {
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

func BotToUser(bot traq.Bot) traq.User {
	user := traq.User{
		Id:          bot.Id,
		Name:        bot.BotUserId,
		DisplayName: "",
		IconFileId:  "",
		Bot:         true,
		State:       BotStateToUserState(bot.State),
		UpdatedAt:   bot.UpdatedAt,
	}
	return user
}

func BotStateToUserState(botState traq.BotState) traq.UserAccountState {
	switch botState {
	case traq.BOTSTATE_deactivated:
		return traq.USERACCOUNTSTATE_deactivated
	case traq.BOTSTATE_active:
		return traq.USERACCOUNTSTATE_active
	}
	return traq.USERACCOUNTSTATE_suspended
}
