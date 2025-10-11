package apiutil

type LocationType string

func (lt LocationType) TypeName() string {
	switch lt {
	case PathLocation:
		return "Path"
	case QueryLocation:
		return "Query"
	case BodyLocation:
		return "Body"
	case HeaderLocation:
		return "Header"
	case DBQueryLocation:
		return "Database Query"
	case UnspecifiedLocationType:
		fallthrough
	default:
		return "Unspecified"
	}
}

func (lt LocationType) Slug() string {
	return string(lt)
}

const UnspecifiedLocationType LocationType = "unspecified"

const (
	PathLocation   LocationType = "path"
	QueryLocation  LocationType = "query"
	BodyLocation   LocationType = "body"
	HeaderLocation LocationType = "header"
)
const (
	DBQueryLocation LocationType = "db_query"
)
