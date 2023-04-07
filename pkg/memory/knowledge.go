package memory

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Knowledge struct {
	ID        string    `json:"id,omitempty" db:"id"`
	Agent     string    `json:"agent,omitempty" db:"agent"`
	User      string    `json:"user,omitempty" db:"user"`
	Subject   string    `json:"subject,omitempty" db:"subject"`
	Predicate string    `json:"predicate,omitempty" db:"predicate"`
	Object    string    `json:"object,omitempty" db:"object"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

func (tidbit *Knowledge) Equal(other *Knowledge) bool {
	timeDifference := tidbit.CreatedAt.Sub(other.CreatedAt)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}
	if timeDifference > time.Second {
		return false
	}

	timeDifference = tidbit.ExpiresAt.Sub(other.ExpiresAt)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}
	if timeDifference > time.Second {
		return false
	}

	return tidbit.ID == other.ID &&
		tidbit.Agent == other.Agent &&
		tidbit.User == other.User &&
		tidbit.Subject == other.Subject &&
		tidbit.Predicate == other.Predicate &&
		tidbit.Object == other.Object
}

func (tidbit *Knowledge) String() string {
	return fmt.Sprintf(
		"%s %s %s",
		tidbit.Subject,
		tidbit.Predicate,
		tidbit.Object,
	)
}

func (tidbit *Knowledge) StringWithExpiration() string {
	var expirationString string
	expiresIn := tidbit.ExpirationToString()
	if expiresIn == "never" {
		expirationString = fmt.Sprintf("(expires %s)", expiresIn)
	} else {
		expirationString = fmt.Sprintf("(expires in %s)", expiresIn)
	}

	return fmt.Sprintf(
		"%s %s %s (%s)",
		tidbit.Subject,
		tidbit.Predicate,
		tidbit.Object,
		expirationString,
	)
}

func (tidbit *Knowledge) IsExpired() bool {
	return time.Now().After(tidbit.ExpiresAt)
}

func (tidbit *Knowledge) ExpiresIn(duration time.Duration) {
	tidbit.ExpiresAt = time.Now().Add(duration)
}

func (tidbit *Knowledge) ExpirationToString() string {
	duration := tidbit.ExpiresAt.Sub(tidbit.CreatedAt)
	var unit string
	var value int

	switch {
	case duration >= time.Hour*24*365*10:
		return "never"
	case duration >= time.Hour*24*7*30:
		value = int(duration / (time.Hour * 24 * 7 * 30))
		unit = pluralize(value, "month")
	case duration >= time.Hour*24*7:
		value = int(duration / (time.Hour * 24 * 7))
		unit = pluralize(value, "week")
	case duration >= time.Hour*24:
		value = int(duration / (time.Hour * 24))
		unit = pluralize(value, "day")
	case duration >= time.Hour:
		value = int(duration / time.Hour)
		unit = pluralize(value, "hour")
	case duration >= time.Minute:
		value = int(duration / time.Minute)
		unit = pluralize(value, "minute")
	default:
		return "less than a minute"
	}

	return fmt.Sprintf("%d %s", value, unit)
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return singular
	}
	return singular + "s"
}

func parseDuration(input string) (time.Duration, error) {
	if strings.ToLower(input) == "never" {
		return time.Hour * 24 * 365 * 100, nil
	}
	// Split the input into the number and the unit
	parts := strings.Split(strings.TrimSpace(input), " ")
	if len(parts) != 2 {
		fmt.Println(">>", parts)
		return 0, fmt.Errorf("invalid duration format: %s", input)
	}

	// Parse the number
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid number in duration: %s", parts[0])
	}

	// Determine the unit of time
	unit := parts[1]
	if unit[len(unit)-1] == 's' {
		unit = unit[:len(unit)-1]
	}

	// Calculate the duration based on the unit
	switch unit {
	case "second", "seconds":
		return time.Duration(num) * time.Second, nil
	case "minute", "minutes":
		return time.Duration(num) * time.Minute, nil
	case "hour", "hours":
		return time.Duration(num) * time.Hour, nil
	case "day", "days":
		return time.Duration(num) * time.Hour * 24, nil
	case "week", "weeks":
		return time.Duration(num) * time.Hour * 24 * 7, nil
	case "month", "months":
		return time.Duration(num) * time.Hour * 24 * 7 * 30, nil
	default:
		return 0, fmt.Errorf("invalid unit of time: %s", unit)
	}
}

type LearnResponse struct {
	Subject   string `json:"subject,omitempty"`
	Predicate string `json:"predicate,omitempty"`
	Object    string `json:"object,omitempty"`
	Expires   string `json:"expires,omitempty"`
}

func ToKnowledge(
	response *LearnResponse,
	agent string,
	user string,
	conversation string,
) (*Knowledge, error) {
	expiresAt, err := parseDuration(response.Expires)
	if err != nil {
		return nil, err
	}

	knowledge := &Knowledge{
		ID:        uuid.New().String(),
		Agent:     agent,
		User:      user,
		CreatedAt: time.Now(),
		Subject:   response.Subject,
		Predicate: response.Predicate,
		Object:    response.Object,
		ExpiresAt: time.Now().Add(expiresAt),
	}

	return knowledge, nil
}
