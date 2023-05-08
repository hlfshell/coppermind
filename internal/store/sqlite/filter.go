package sqlite

import (
	"fmt"
	"strings"

	"github.com/hlfshell/coppermind/internal/store"
)

func filterToQueryParams(filter store.Filter) (string, []interface{}, error) {
	query := strings.Builder{}
	params := []interface{}{}
	for i, fc := range filter.Attributes {
		// Add an AND if this is not the first column
		// being added
		if i != 0 {
			query.WriteString(" AND ")
		}

		// Switch state for the operations allowed
		switch fc.Operation {
		case store.EQ:
			query.WriteString(fmt.Sprintf("%s = ?", fc.Attribute))
			params = append(params, fc.Value)
		case store.NEQ:
			query.WriteString(fmt.Sprintf("%s != ?", fc.Attribute))
			params = append(params, fc.Value)
		case store.GT:
			query.WriteString(fmt.Sprintf("%s > ?", fc.Attribute))
			params = append(params, fc.Value)
		case store.LT:
			query.WriteString(fmt.Sprintf("%s < ?", fc.Attribute))
			params = append(params, fc.Value)
		case store.GTE:
			query.WriteString(fmt.Sprintf("%s >= ?", fc.Attribute))
			params = append(params, fc.Value)
		case store.LTE:
			query.WriteString(fmt.Sprintf("%s <= ?", fc.Attribute))
			params = append(params, fc.Value)
		case store.IN:
			query.WriteString(fmt.Sprintf("%s IN (?)", fc.Attribute))
			params = append(params, fc.Value)
		default:
			return "", nil, fmt.Errorf("invalid operation %s", fc.Operation)
		}
	}

	return query.String(), params, nil
}
