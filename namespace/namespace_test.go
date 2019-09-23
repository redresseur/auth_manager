package namespace

import (
	"encoding/json"
	"github.com/redresseur/auth_manager/namespace/acl_condition"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	parentNamespace *RestFulAuthorNamespace
)

func TestNewNameSpace(t *testing.T) {
	p := acl_condition.AND(acl_condition.BASE("PUT DATA"),
		acl_condition.BASE("GET DATA"))

	cond := acl_condition.NewAclCondition(p)
	parentNamespace = NewNameSpace("/", WithDefaultCond(cond))

	t.Logf("parentNamespace: %v", parentNamespace)
}

func TestAddSubNameSpace(t *testing.T) {
	TestNewNameSpace(t)

	childNamespace, err := AddSubNameSpace(parentNamespace, "test/first")
	if !assert.NoError(t, err) {
		t.SkipNow()
	} else {
		t.Logf("childNamespace: %v", childNamespace)
	}
}

func TestMarshal(t *testing.T) {
	TestAddSubNameSpace(t)
	if data, err := json.Marshal(parentNamespace); err != nil {
		t.Errorf("marshal parentNamespace: %v", err)
	} else {
		t.Log(string(data))
		tn := &RestFulAuthorNamespace{
			SrcNamespace: &source_namespace{
				ConditionDefault: acl_condition.NewAclCondition(nil),
			},
		}
		if err := json.Unmarshal(data, tn); err != nil {
			t.Errorf("unmarshal parentNamespace: %v", err)
		} else {
			t.Logf("%++v", tn)
		}
	}
}

func TestNameSpace(t *testing.T) {
	TestAddSubNameSpace(t)
	t.Log(NameSpace(parentNamespace, "test/first"))
}
