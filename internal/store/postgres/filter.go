package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hlfshell/coppermind/internal/store"
)

func filterToQueryParams(filter store.Filter) (string, []interface{}, error) {
	query := strings.Builder{}
	params := []interface{}{}
	placeholderCount := 1

	for i, fc := range filter.Attributes {
		// If the column is user, replace it with userId
		if fc.Attribute == "user" {
			fc.Attribute = "userId"
		}

		// Add an AND if this is not the first column
		// being added
		if i != 0 {
			query.WriteString(" AND ")
		}

		// Switch state for the operations allowed
		switch fc.Operation {
		case store.EQ:
			query.WriteString(fmt.Sprintf("%s = $%d", fc.Attribute, placeholderCount))
			placeholderCount++
			params = append(params, fc.Value)
		case store.NEQ:
			query.WriteString(fmt.Sprintf("%s != $%d", fc.Attribute, placeholderCount))
			placeholderCount++
			params = append(params, fc.Value)
		case store.GT:
			query.WriteString(fmt.Sprintf("%s > $%d", fc.Attribute, placeholderCount))
			placeholderCount++
			params = append(params, fc.Value)
		case store.LT:
			query.WriteString(fmt.Sprintf("%s < $%d", fc.Attribute, placeholderCount))
			placeholderCount++
			params = append(params, fc.Value)
		case store.GTE:
			query.WriteString(fmt.Sprintf("%s >= $%d", fc.Attribute, placeholderCount))
			placeholderCount++
			params = append(params, fc.Value)
		case store.LTE:
			query.WriteString(fmt.Sprintf("%s <= $%d", fc.Attribute, placeholderCount))
			placeholderCount++
			params = append(params, fc.Value)
		case store.IN:
			// You can't just pass an array in as a param for sqlite
			// like you can postgres, so we have to do some additional
			// massaging. Thus we have to return a placeholder for
			// each item *and* pass in each item individually

			// Convert our interface{} to a []interface{} since it's
			// assumed that's what was passed to us. We'll use reflect
			rv := reflect.ValueOf(fc.Value)
			if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
				return "", nil, fmt.Errorf("invalid value for IN operation - expected a slice type")
			}

			// Create our placeholder string and params array
			placeholder := strings.Builder{}
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					placeholder.WriteString(", ")
				}
				placeholder.WriteString(fmt.Sprintf("$%d", placeholderCount))
				placeholderCount++
				params = append(params, rv.Index(i).Interface())
			}

			query.WriteString(fmt.Sprintf("%s IN (%s)", fc.Attribute, placeholder.String()))
		default:
			return "", nil, fmt.Errorf("invalid operation %s", fc.Operation)
		}
	}

	return query.String(), params, nil
}
