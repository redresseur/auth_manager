package auth_manager

import (
	"github.com/gin-gonic/gin"
	"github.com/redresseur/auth_manager/namespace"
)

var permission = func(ctx *gin.Context) (interface{}, error) {
	return nil, nil
}

func RegistryPermission(pm func(ctx *gin.Context) (interface{}, error) )  {
	permission = pm
}

func CheckAuthority(ctx *gin.Context, namespaces []*namespace.RestFulAuthorNamespace) error {
	p, err := permission(ctx)
	if err != nil{
		return err
	}

	for _, sp := range namespaces{
		if err = namespace.AuthorityCheck(sp, p); err != nil{
			return err
		}
	}

	return nil
}