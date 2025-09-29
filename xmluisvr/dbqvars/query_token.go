package dbqvars

import (
	"sort"
)

type QueryToken struct {
	Name  Parameter // logical name: e.g., "path.accountId" or "body.items.0.id"
	Index int       // assigned parameter index (1-based)
	Start int       // byte offset start in original SQL
	End   int       // byte offset end (exclusive)
	Raw   string    // full token, e.g. "{user.id}"
}

type QueryTokens []QueryToken

func (qts QueryTokens) Parameters() (names []Parameter) {
	names = make([]Parameter, len(qts))
	// Ensure the parameters are ordered by Index
	sort.Slice(qts, func(i, j int) bool {
		return qts[i].Index < qts[j].Index
	})
	for i, sp := range qts {
		names[i] = sp.Name
	}
	return names
}

// GetValues returns the values for SQL parameters
// TODO This logic should be tied to both the map type and the json type
//
//	and then merged where it is currently being used. It is too tightly coupled as is.
//func (qts QueryTokens) GetValues(m map[common.Identifier]any, jsonValues common.JSONBytes) (values []any, err error) {
//	var errs []error
//	values = make([]any, len(qts))
//	// Ensure the parameters are ordered by Index
//	sort.Slice(qts, func(i, j int) bool {
//		return qts[i].Index < qts[j].Index
//	})
//	jve := jsonutil.NewValueExtractor()
//	err = jve.ParseBytes(jsonValues)
//	if err != nil {
//		err = errors.Join(jsonutil.ErrFailedParsingJSONForDBQueryParameters, jve.ErrorString(), err)
//		goto end
//	}
//	for i, param := range qts {
//		var value any
//		var ok bool
//		value, ok = m[common.Identifier(param.Name)]
//		if ok {
//			values[i] = value
//			continue
//		}
//		value, err = jve.ExtractValue(common.Selector(param.Name))
//		if err != nil {
//			errs = append(errs, errors.Join(jsonutil.ErrUndefinedSQLParameter,
//				fmt.Errorf("sql_parameter=%s", param.Name),
//				fmt.Errorf("positon=%d", param.Index),
//			))
//			continue
//		}
//		values[i] = value
//	}
//end:
//	return values, err
//}
