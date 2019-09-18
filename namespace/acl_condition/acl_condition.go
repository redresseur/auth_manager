package acl_condition

import (
	"github.com/redresseur/auth_manager/common/err"
)

type aclCondition struct {
	policy *ApiPolicy
}

type AclPermission struct {
	AuthList map[string]bool `json:"authList"`
}

func (acc *aclCondition) Check(Permission interface{}) error {
	ap, ok := Permission.(*AclPermission)
	if !ok {
		return err.ErrPermissionNotValid
	}

	if !acc.policy.check(ap.AuthList) {
		return err.ErrPermissionCheckFailure
	}

	return nil
}
