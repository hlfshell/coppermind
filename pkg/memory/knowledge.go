package memory

import (
	"fmt"
	"time"

	"github.com/drewlanenga/govector"
)

type Knowledge struct {
	ID     string `json:"id,omitempty" db:"id"`
	Agent  string `json:"agent,omitempty" db:"agent"`
	User   string `json:"user,omitempty" db:"user"`
	Source string `json:"source,omitempty" db:"source"`

	// Metadata is source specific. For example, a document may have
	// additional information such as file location, an associated
	// URL, author information, etc. IT's information we don't expect
	// to be required to search by, but may be necessary upon recall
	Metadata map[string]string `json:"metadata,omitempty" db:"metadata"`

	Content string          `json:"content,omitempty" db:"content"`
	Vector  govector.Vector `json:"vector,omitempty" db:"vector"`

	CreatedAt    time.Time `json:"created_at,omitempty" db:"created_at"`
	LastUtilized time.Time `json:"last_utilized,omitempty" db:"last_utilized"`
}

func (tidbit *Knowledge) Equal(other *Knowledge) bool {
	timeDifference := tidbit.CreatedAt.Sub(other.CreatedAt)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}
	if timeDifference > time.Second {
		return false
	}

	timeDifference = tidbit.LastUtilized.Sub(other.LastUtilized)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}
	if timeDifference > time.Second {
		return false
	}

	for key, value := range tidbit.Metadata {
		otherValue, ok := other.Metadata[key]
		if !ok || value != otherValue {
			return false
		}
	}
	// Ensure that there are no extraneous keys
	if len(tidbit.Metadata) != len(other.Metadata) {
		return false
	}

	if len(tidbit.Vector) != len(other.Vector) {
		return false
	}
	for i, value := range tidbit.Vector {
		if value != other.Vector[i] {
			return false
		}
	}

	return tidbit.ID == other.ID &&
		tidbit.Agent == other.Agent &&
		tidbit.User == other.User &&
		tidbit.Content == other.Content
}

func (tidbit *Knowledge) String() string {
	return fmt.Sprintf(
		"%s | %s",
		tidbit.ID,
		tidbit.Content[0:50],
	)
}
