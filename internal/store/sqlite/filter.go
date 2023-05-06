package sqlite

import (
	"fmt"
	"strings"

	"github.com/hlfshell/coppermind/internal/store"
)

func filterToQueryParams(filter store.Filter) ([]string, []interface{}, error) {
	query := strings.Builder{}
	columns := []string{}
	params := []interface{}{}
	for i, fc := range filter.Columns {
		// Add an AND if this is not the first column
		// being added
		if i != 0 {
			query.WriteString(" AND ")
		}

		// Switch state for the operations allowed
		switch fc.Operation {
		case store.EQ:
			query.WriteString(fmt.Sprintf("%s = ?", fc.Column))
			columns = append(columns, fc.Column)
			params = append(params, fc.Value)
		case store.NEQ:
			query.WriteString(fmt.Sprintf("%s != ?", fc.Column))
			columns = append(columns, fc.Column)
			params = append(params, fc.Value)
		case store.GT:
			query.WriteString(fmt.Sprintf("%s > ?", fc.Column))
			columns = append(columns, fc.Column)
			params = append(params, fc.Value)
		case store.LT:
			query.WriteString(fmt.Sprintf("%s < ?", fc.Column))
			columns = append(columns, fc.Column)
			params = append(params, fc.Value)
		case store.GTE:
			query.WriteString(fmt.Sprintf("%s >= ?", fc.Column))
			columns = append(columns, fc.Column)
			params = append(params, fc.Value)
		case store.LTE:
			query.WriteString(fmt.Sprintf("%s <= ?", fc.Column))
			columns = append(columns, fc.Column)
			params = append(params, fc.Value)
		case store.IN:
			query.WriteString(fmt.Sprintf("%s IN (?)", fc.Column))
			columns = append(columns, fc.Column)
			params = append(params, fc.Value)
		default:
			return nil, nil, fmt.Errorf("invalid operation %s", fc.Operation)
		}
	}

	return columns, params, nil
}