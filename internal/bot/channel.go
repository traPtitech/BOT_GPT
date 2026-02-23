package bot

import (
	"context"
	"time"

	"github.com/motoki317/sc"
)

func getChannelPathInternal(ctx context.Context, channelID string) (string, error) {
	bot := GetBot()

	path, _, err := bot.API().ChannelAPI.GetChannelPath(ctx, channelID).Execute()
	if err != nil {
		return "", err
	}

	return path.Path, nil
}

var channelPathCache = sc.NewMust(getChannelPathInternal, time.Hour, time.Hour, nil)

func GetChannelPath(channelID string) (string, error) {
	return channelPathCache.Get(context.Background(), channelID)
}
