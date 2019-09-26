package namespace

import (
	"encoding/json"
	"github.com/redresseur/auth_manager/namespace/acl_condition"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

var (
	parentNamespace *RestFulAuthorNamespace
)

func TestNewNameSpace(t *testing.T) {
	p := acl_condition.AND(acl_condition.BASE("PUT DATA"),
		acl_condition.BASE("GET DATA"))

	cond := acl_condition.NewAclCondition(p)
	parentNamespace = NewNameSpace("/", nil, WithDefaultCond(cond))

	t.Logf("parentNamespace: %v", parentNamespace)
}

func TestAddSubNameSpace(t *testing.T) {
	TestNewNameSpace(t)

	childNamespace, err := AddSubNameSpace(parentNamespace, "/test/first")
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
	t.Log(NameSpace(parentNamespace, "/test/first"))
}

func TestParamAll(t *testing.T) {
	parentNamespace = NewNameSpace("/", nil)
	AddSubNameSpace(parentNamespace, "/test/*/all")
	t.Log(NameSpace(parentNamespace, "/test/aa/all"))
}

func TestCatchParam(t *testing.T) {
	ok, param := checkParam("test{test}")
	assert.Equal(t, false, ok, "test{test} is not ok")

	ok, param = checkParam("{test}{test}")
	assert.Equal(t, false, ok, "{test}{test} is not ok")

	ok, param = checkParam("{test}")
	assert.Equal(t, true, ok, "{test} is not ok")

	t.Logf("param is %s", param)
}

func TestCatchParam1(t *testing.T) {
	// rc, err := regexp.Compile(`/(\w[^\/]+){*}?([^/:]+)/`)
	rc, err := regexp.Compile(`(\{.[a-zA-Z\_0-9]+\})`)

	if !assert.NoError(t, err) {
		t.SkipNow()
	}

	t.Log(rc.FindAllString("/person/{id}/name/{action}/test", -1))
}

func TestParam(t *testing.T) {
	rc, err := regexp.Compile(`(b*)`)
	if !assert.NoError(t, err) {
		t.SkipNow()
	}

	t.Log(rc.FindAllString("banga", -1))
}

func TestPathParse(t *testing.T) {
	t.Log(PathParse("/person/{id}/name/{action}/test"))
}
