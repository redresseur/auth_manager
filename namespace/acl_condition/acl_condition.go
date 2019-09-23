package acl_condition

import (
	"github.com/redresseur/auth_manager/common/err"
)

type aclCondition struct {
	Policy *ApiPolicy `json:"policy"`
}

type AclPermission struct {
	AuthList map[string]bool `json:"authList"`
}

func NewAclCondition(policy *ApiPolicy) *aclCondition {
	return &aclCondition{Policy: policy}
}

func (acc *aclCondition) Check(Permission interface{}) error {
	ap, ok := Permission.(*AclPermission)
	if !ok {
		return err.ErrPermissionNotValid
	}

	if !acc.Policy.check(ap.AuthList) {
		return err.ErrPermissionCheckFailure
	}

	return nil
}
