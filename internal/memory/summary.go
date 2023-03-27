package memory

import (
	"strings"
	"time"

	"github.com/wissance/stringFormatter"
)

type Summary struct {
	ID           string
	Agent        string
	Conversation string
	Keywords     []string
	Brief        string
	User         string
	CreatedAt    time.Time
}

func (summary *Summary) KeywordsToString() string {
	return strings.Join(summary.Keywords, ",")
}

func (summary *Summary) StringToKeywords(input string) {
	summary.Keywords = strings.Split(input, ",")
}

func (summary *Summary) String() string {
	return stringFormatter.Format(
		"{0} | {1}",
		summary.KeywordsToString(),
		summary.Brief,
	)
}
