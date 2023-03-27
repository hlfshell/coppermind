package memory

import "time"

type Knowledge struct {
	Subject   string
	Predicate string
	Object    string
	CreatedAt time.Time
	ExpiresIn time.Time
}
