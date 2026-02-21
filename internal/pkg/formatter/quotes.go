package formatter

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/traPtitech/BOT_GPT/internal/bot"
	"github.com/traPtitech/go-traq"
)

const quoteRegexStr = `\bhttps://q\.trap\.jp/messages/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\b`

var quoteRegex = regexp.MustCompile(quoteRegexStr)

var allowingPrefixes = []string{"/event", "/general", "/random", "/services", "/team/SysAd"}

func isChannelAllowingQuotes(channelID string) (bool, error) {
	channelPath, err := bot.GetChannelPath(channelID)
	if err != nil {
		return false, err
	}

	for _, prefix := range allowingPrefixes {
		if strings.HasPrefix(channelPath, prefix) {
			return true, nil
		}
	}

	return false, nil
}

func isUserAllowingQuotes(userID string, messageUserID string) (bool, error) {
	if userID == messageUserID {
		return true, nil
	}

	messageUser, err := bot.GetUser(messageUserID)
	if err != nil {
		return false, err
	}

	if messageUser.Bot {
		return true, nil
	}

	return false, nil
}

func getQuoteMarkdown(message *traq.Message) (string, error) {
	user, err := bot.GetUser(message.UserId)
	if err != nil {
		return "", err
	}

	return "> " + user.Name + ":\n> " + strings.ReplaceAll(message.Content, "\n", "\n> "), nil
}

const maxQuoteLength = 10000

func FormatQuotedMessage(userID string, content string) (string, error) {
	matches := quoteRegex.FindAllSubmatch([]byte(content), len(content))
	messageIDs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		messageID := string(match[1])
		messageIDs = append(messageIDs, messageID)
	}

	var formattedContent strings.Builder
	formattedContent.WriteString(quoteRegex.ReplaceAllString(content, ""))

	for _, messageID := range messageIDs {
		message := bot.GetMessage(messageID)
		if message == nil {
			continue
		}

		if utf8.RuneCountInString(message.Content) > maxQuoteLength {
			runes := []rune(message.Content)
			message.Content = string(runes[:maxQuoteLength]) + "(以下略)"
		}

		channelAllowed, err := isChannelAllowingQuotes(message.ChannelId)
		if err != nil {
			return "", err
		}
		userAllowed, err := isUserAllowingQuotes(userID, message.UserId)
		if err != nil {
			return "", err
		}
		if !channelAllowed && !userAllowed {
			continue
		}

		quote, err := getQuoteMarkdown(message)
		if err != nil {
			return "", err
		}
		formattedContent.WriteString("\n\n" + quote)
	}
	return formattedContent.String(), nil
}
