package bot

import (
	"context"
	"time"

	"github.com/motoki317/sc"
	"github.com/traPtitech/go-traq"
)

func getUserInternal(ctx context.Context, userID string) (*traq.UserDetail, error) {
	bot := GetBot()

	user, _, err := bot.API().UserAPI.GetUser(ctx, userID).Execute()
	if err != nil {
		return nil, err
	}

	return user, nil
}

var userCache = sc.NewMust(getUserInternal, time.Hour, time.Hour, nil)

func GetUser(userID string) (*traq.UserDetail, error) {
	return userCache.Get(context.Background(), userID)
}
