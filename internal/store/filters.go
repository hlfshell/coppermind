package store

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
	EQ  = "="
	NEQ = "!="
	GT  = ">"
	LT  = "<"
	GTE = ">="
	LTE = "<="
	IN  = "in"
)

/*
FilterColumn is a generic column identifier that can
be utilized in search query filters to determine if:

1 - the column is being utilized in the search

2 - what type of comparison is being utilized (ie,
equality, greater than, less than, etc)

3 - what value(s) is being utilized in the comparison

This is done generically so it can be adapted to multiple
store types
*/
type FilterColumn struct {
	Column    string
	Operation string
	Value     interface{}
}

/*
Filter is a generic filter that can be utilized in search
query filters to determine how to construct a complex query
for the given object.

Pagination is expected to occur within the store if necessary.

The Limit is optional - if <= 0 it is to be ignored.
*/
type Filter struct {
	Columns []*FilterColumn
	Limit   int
}

// Quick check to see if the filter is empty, suggestings a "select all"
// equivalent
func (filter *Filter) Empty() bool {
	return len(filter.Columns) == 0
}
