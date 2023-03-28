package memory

import (
	"strings"
	"time"

	"github.com/wissance/stringFormatter"
)

type Summary struct {
	ID           string    `json:"id,omitempty" db:"id"`
	Agent        string    `json:"agent,omitempty" db:"agent"`
	Conversation string    `json:"conversation,omitempty" db:"conversation"`
	Keywords     []string  `json:"keywords,omitempty" db:"keywords"`
	Summary      string    `json:"summary,omitempty" db:"summary"`
	User         string    `json:"user,omitempty" db:"user"`
	CreatedAt    time.Time `json:"created_at,omitempty" db:"created_at"`
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
		summary.Summary,
	)
}
