package err

import "errors"

var (
	ErrPermissionNotValid     = errors.New("the permission is invalid")
	ErrPermissionCheckFailure = errors.New("check the permission, no passed")
)
