package store

import (
	"fmt"
	"strings"
	"time"
)

/*
These are the expected operations for the filters:
1. = - equality
2. != - not equal
3. > - greater than
4. < - less than
5. >= - greater than or equal
6. <= - less than or equal
7. in - in a list of values

It is expected that it is supported in some manner
for all stores.
*/
const (
	EQ  = "=="
	NEQ = "!="
	GT  = ">"
	LT  = "<"
	GTE = ">="
	LTE = "<="
	IN  = "in"
)

/*
FilterAttribute is a generic column identifier that can
be utilized in search query filters to determine if:

1 - the column is being utilized in the search

2 - what type of comparison is being utilized (ie,
equality, greater than, less than, etc)

3 - what value(s) is being utilized in the comparison

This is done generically so it can be adapted to multiple
store types
*/
type FilterAttribute struct {
	Attribute string
	Operation string
	Value     interface{}
}

func (attribute *FilterAttribute) String() string {
	return fmt.Sprintf("%s %s %v", attribute.Attribute, attribute.Operation, attribute.Value)
}

type OrderBy struct {
	Attribute string
	Ascending bool
}

// Nil checks to see if we're essentially "unset" on the order by
func (orderBy *OrderBy) Nil() bool {
	return orderBy.Attribute == ""
}

/*
Filter is a generic filter that can be utilized in search
query filters to determine how to construct a complex query
for the given object.

Pagination is expected to occur within the store if necessary.

The Limit is optional - if <= 0 it is to be ignored.
*/
type Filter struct {
	Attributes []*FilterAttribute
	OrderBy    OrderBy
	Limit      int
}

// Quick check to see if the filter is empty, suggestings a "select all"
// equivalent
func (filter *Filter) Empty() bool {
	return len(filter.Attributes) == 0
}

type FilterString struct {
	Operation string
	Value     string
}

func (filter *FilterString) Valid() bool {
	return filter.Operation == EQ ||
		filter.Operation == NEQ ||
		filter.Operation == IN
}

// In converts a comma delimited list of values to a []string value
// to support IN equivalent operations
func (filter *FilterString) In() []string {
	values := strings.Split(filter.Value, ",")
	for i, value := range values {
		values[i] = strings.TrimSpace(value)
	}
	return values
}

func (filter *FilterString) String() string {
	if filter.Operation == IN {
		return fmt.Sprintf("%s %s", filter.Operation, filter.In())
	} else {
		return fmt.Sprintf("%s %s", filter.Operation, filter.Value)
	}
}

type FilterTime struct {
	Operation string
	Value     time.Time
}

func (filter *FilterTime) Valid() bool {
	return (filter.Operation == EQ ||
		filter.Operation == NEQ ||
		filter.Operation == GT ||
		filter.Operation == LT ||
		filter.Operation == GTE ||
		filter.Operation == LTE) &&
		!filter.Value.IsZero()
}

func (filter *FilterTime) String() string {
	return fmt.Sprintf("%s %s", filter.Operation, filter.Value)
}

// ====================
// Model specific filters
// ====================

type MessageFilter struct {
	ID           *FilterString `json:"id,omitempty"`
	Conversation *FilterString `json:"conversation,omitempty"`
	User         *FilterString `json:"user,omitempty"`
	Agent        *FilterString `json:"agent,omitempty"`
	From         *FilterString `json:"from,omitempty"`
	CreatedAt    *FilterTime   `json:"created_at,omitempty"`
}

func (filter *MessageFilter) Empty() bool {
	return filter.ID == nil &&
		filter.Conversation == nil &&
		filter.User == nil &&
		filter.Agent == nil &&
		filter.From == nil &&
		filter.CreatedAt == nil
}

func (filter *MessageFilter) String() string {
	return fmt.Sprintf("MessageFilter: \n\t%s \n\t%s \n\t%s \n\t%s \n\t%s \n\t%s",
		filter.ID,
		filter.Conversation,
		filter.User,
		filter.Agent,
		filter.From,
		filter.CreatedAt)
}

type ConversationFilter struct {
	ID        *FilterString `json:"id,omitempty"`
	Agent     *FilterString `json:"agent,omitempty"`
	User      *FilterString `json:"user,omitempty"`
	CreatedAt *FilterTime   `json:"created_at,omitempty"`
}

func (filter *ConversationFilter) Empty() bool {
	return filter.ID == nil &&
		filter.Agent == nil &&
		filter.User == nil &&
		filter.CreatedAt == nil
}

func (filter *ConversationFilter) String() string {
	return fmt.Sprintf("ConversationFilter: \n\t%s \n\t%s \n\t%s \n\t%s",
		filter.ID,
		filter.Agent,
		filter.User,
		filter.CreatedAt)
}

type SummaryFilter struct {
	ID                    *FilterString `json:"id,omitempty"`
	Agent                 *FilterString `json:"agent,omitempty"`
	User                  *FilterString `json:"user,omitempty"`
	Conversaton           *FilterString `json:"conversation,omitempty"`
	UpdatedAt             *FilterTime   `json:"updated_at,omitempty"`
	ConversationStartedAt *FilterTime   `json:"conversation_started_at,omitempty"`
}

func (filter *SummaryFilter) Empty() bool {
	return filter.ID == nil &&
		filter.Agent == nil &&
		filter.User == nil &&
		filter.Conversaton == nil &&
		filter.UpdatedAt == nil &&
		filter.ConversationStartedAt == nil
}

func (filter *SummaryFilter) String() string {
	return fmt.Sprintf("SummaryFilter: \n\t%s \n\t%s \n\t%s \n\t%s \n\t%s \n\t%s",
		filter.ID,
		filter.Agent,
		filter.User,
		filter.Conversaton,
		filter.UpdatedAt,
		filter.ConversationStartedAt)
}

type KnowledgeFilter struct {
	ID           *FilterString `json:"id,omitempty"`
	Agent        *FilterString `json:"agent,omitempty"`
	User         *FilterString `json:"user,omitempty"`
	Source       *FilterString `json:"source,omitempty"`
	CreatedAt    *FilterTime   `json:"created_at,omitempty"`
	LastUtilized *FilterTime   `json:"last_accessed,omitempty"`

	OldestFirst bool `json:"oldest_first,omitempty"`
	Limit       int  `json:"limit,omitempty"`
}

func (filter *KnowledgeFilter) Empty() bool {
	return filter.ID == nil &&
		filter.Agent == nil &&
		filter.User == nil &&
		filter.Source == nil &&
		filter.CreatedAt == nil &&
		filter.LastUtilized == nil
}

func (filter *KnowledgeFilter) String() string {
	return fmt.Sprintf("KnowledgeFilter: \n\t%s \n\t%s \n\t%s \n\t%s \n\t%s \n\t%s",
		filter.ID,
		filter.Agent,
		filter.User,
		filter.Source,
		filter.CreatedAt,
		filter.LastUtilized)
}
