package formatter

import "regexp"

var embedRegex = regexp.MustCompile(`!\{"type":"(\w+?)","raw":"(.+?)","id":"[a-z0-9-]+?"\}`)

func FormatEmbeds(content string) string {
	return embedRegex.ReplaceAllString(content, "$2")
}
