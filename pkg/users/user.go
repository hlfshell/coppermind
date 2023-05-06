package users

import (
	"time"
)

type User struct {
	ID        string    `json:"id,omitempty" db:"id"`
	Name      string    `json:"name,omitempty" db:"name"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

func (user *User) Equal(other *User) bool {
	timeDiffereneceCreatedAt := user.CreatedAt.Sub(other.CreatedAt)
	timeDifferenceUpdatedAt := user.UpdatedAt.Sub(other.UpdatedAt)

	if timeDiffereneceCreatedAt < 0 {
		timeDiffereneceCreatedAt = -timeDiffereneceCreatedAt
	}
	if timeDifferenceUpdatedAt < 0 {
		timeDifferenceUpdatedAt = -timeDifferenceUpdatedAt
	}

	return user.ID == other.ID &&
		user.Name == other.Name &&
		timeDiffereneceCreatedAt < time.Second &&
		timeDifferenceUpdatedAt < time.Second
}
