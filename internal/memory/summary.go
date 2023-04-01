package memory

import (
	"strings"
	"time"

	"github.com/wissance/stringFormatter"
)

type Summary struct {
	ID                    string    `json:"id,omitempty" db:"id"`
	Agent                 string    `json:"agent,omitempty" db:"agent"`
	Conversation          string    `json:"conversation,omitempty" db:"conversation"`
	Keywords              []string  `json:"keywords,omitempty" db:"keywords"`
	Summary               string    `json:"summary,omitempty" db:"summary"`
	User                  string    `json:"user,omitempty" db:"user"`
	UpdatedAt             time.Time `json:"updated_at,omitempty" db:"updated_at"`
	ConversationStartedAt time.Time `json:"conversation_started_at,omitempty" db:"conversation_started_at"`
}

func (summary *Summary) Equal(other *Summary) bool {
	timeDifference := summary.UpdatedAt.Sub(other.UpdatedAt)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}
	if timeDifference > time.Second {
		return false
	}

	keywordCheck := map[string]bool{}
	for _, keyword := range summary.Keywords {
		keywordCheck[keyword] = false
	}
	for _, keyword := range other.Keywords {
		if _, ok := keywordCheck[keyword]; ok {
			keywordCheck[keyword] = ok
		} else {
			return false
		}
	}
	for _, keyword := range summary.Keywords {
		if !keywordCheck[keyword] {
			return false
		}
	}

	return summary.ID == other.ID &&
		summary.Agent == other.Agent &&
		summary.User == other.User &&
		summary.Conversation == other.Conversation &&
		summary.Summary == other.Summary

}

func (summary *Summary) KeywordsToString() string {
	return strings.Join(summary.Keywords, ",")
}

func (summary *Summary) StringToKeywords(input string) {
	summary.Keywords = strings.Split(input, ",")
}

func (summary *Summary) StringWithConversation() string {
	return stringFormatter.Format(
		"{0} | {1} | {2} | {3}",
		summary.User,
		summary.ConversationStartedAt.Format("January 2nd 06 15:04"),
		summary.KeywordsToString(),
		summary.Summary,
	)
}
