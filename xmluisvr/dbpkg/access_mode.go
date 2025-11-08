package dbpkg

import (
	"github.com/xmlui-org/localsvr/xmluisvr/dbqvars"
)

// Constants for AccessMode values
// IMPORTANT: Number values of the AccessModes are critical to the algorithm
const (
	UnspecifiedAccessMode = AccessMode(dbqvars.UnspecifiedDBAccessMode)
	ReadOnlyMode          = AccessMode(dbqvars.DBReadOnlyMode)
	ReadWriteMode         = AccessMode(dbqvars.DBReadWriteMode)
	AdminMode             = AccessMode(dbqvars.DBAdminMode)
	SuperAdminMode        = AccessMode(dbqvars.DBSuperAdminMode)
)

type AccessMode int

func (m AccessMode) Name() string {
	switch m {
	case SuperAdminMode:
		return "Super Admin"
	case AdminMode:
		return "Admin"
	case ReadWriteMode:
		return "Read-Write"
	case ReadOnlyMode:
		return "Read-Only"
	case UnspecifiedAccessMode:
		fallthrough
	default:
		return "Unspecified"
	}
}

func ParseAccessMode(mode int) (am AccessMode, err error) {
	// TODO Validate
	return AccessMode(mode), err
}
