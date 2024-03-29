package namespace

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

const (
	ParamAll  = `*`
	Separator = `/`
)

// 资源访问的条件
type Condition interface {
	// 检查许可
	Check(Permission interface{}) error
}

type EmptyCondition struct {
}

func (*EmptyCondition) Check(Permission interface{}) error {
	return nil
}

// /aaa/bbbb/cccc
// /aaa/{xx}/cccc
type source_namespace struct {

	// 是否為可變的
	Variable bool `json:"variable"`

	// 參數名稱
	ParamName string `json:"variableName"`

	// 讀條件集合，每一個資源請求對應的條件
	// 對應 GET/
	ConditionRead Condition `json:"conditionRead"`

	// 写条件集合
	// 對應POST
	ConditionWrite Condition `json:"conditionWrite"`

	// 删除条件集合
	// 對應 DELETE
	ConditionDelete Condition `json:"conditionDelete"`

	// 更新条件集合
	// 對應 PUT
	ConditionUpdate Condition `json:"conditionUpdate"`

	// 默认条件集合
	// 默認條件
	ConditionDefault Condition `json:"conditionDefault"`

	l sync.Mutex `json:"_"`
}

type RestFulAuthorNamespace struct {
	// 空间名称
	Name string `json:"name"`

	SrcNamespace *source_namespace `json:"srcNamespace"`
	// 子空間
	SubNamespaces map[string]*RestFulAuthorNamespace `json:"subNamespaces"`

	Parent *RestFulAuthorNamespace
}

type CondsOp func(namespace *source_namespace)

func WithReadCond(condition Condition) CondsOp {
	return func(namespace *source_namespace) {
		namespace.ConditionRead = condition
	}
}
func WithWriteCond(condition Condition) CondsOp {
	return func(namespace *source_namespace) {
		namespace.ConditionWrite = condition
	}
}

func WithUpdateCond(condition Condition) CondsOp {
	return func(namespace *source_namespace) {
		namespace.ConditionUpdate = condition
	}
}

func WithDeleteCond(condition Condition) CondsOp {
	return func(namespace *source_namespace) {
		namespace.ConditionDelete = condition
	}
}

func WithDefaultCond(condition Condition) CondsOp {
	return func(namespace *source_namespace) {
		namespace.ConditionDefault = condition
	}
}

func UpdateCondition(namespace *RestFulAuthorNamespace, ops ...CondsOp) error {
	for _, op := range ops {
		op(namespace.SrcNamespace)
	}
	return nil
}

func checkParam(str string) (bool, string) {
	rc, _ := regexp.Compile(`(\{.[a-zA-Z\_0-9]+\})`)
	res := rc.FindAllString(str, -1)
	if len(res) != 1 {
		return false, ""
	}

	param := res[0]
	paramLen := len(param)
	if paramLen != len(str) {
		return false, ""
	}

	return true, param[1 : paramLen-1]
}

func ReverseFind(sp *RestFulAuthorNamespace, name string) (*RestFulAuthorNamespace, error) {
	if sp.Name == name {
		return sp, nil
	}

	if sp.Parent == nil {
		return nil, errors.New("Namespace Not Found")
	}

	return ReverseFind(sp.Parent, name)
}

func PathParse(path string) []string {
	if len(path) == 0 {
		return []string{}
	}

	res := []string{}
	sub := []rune{}
	for _, c := range path {
		if c == 0x2f {
			if len(sub) > 0 {
				res = append(res, string(sub))
				sub = sub[:0]
			}
			res = append(res, Separator)
		} else {
			sub = append(sub, c)
		}
	}

	if len(sub) > 0 {
		res = append(res, string(sub))
	}
	return res
}

// path: 資源路徑
func AddSubNameSpace(parent *RestFulAuthorNamespace, path string) (*RestFulAuthorNamespace, error) {
	var tmp *RestFulAuthorNamespace

	subs := PathParse(path)
	if len(subs) == 0 {
		return nil, errors.New("the path is invalid")
	}

	if subs[0] != Separator && parent.Name != Separator {
		parent, _ = AddSubNameSpace(parent, Separator)
	} else if subs[0] == Separator && parent.Name == Separator {
		subs = subs[1:]
	}

	for _, sub := range subs {
		if sub == "" {
			continue
		}

		// 检查是否已经存在
		if _, ok := parent.SubNamespaces[sub]; ok {
			parent = parent.SubNamespaces[sub]
			tmp = parent
			continue
		}

		// 检查字段是否为可变量
		isParam, param := checkParam(sub)
		tmp = NewNameSpace(sub, parent)
		tmp.SrcNamespace.Variable = isParam
		tmp.SrcNamespace.ParamName = param

		parent.SubNamespaces[sub] = tmp
		parent = tmp
	}

	return parent, nil
}

func NewNameSpace(name string, parent *RestFulAuthorNamespace, ops ...func(namespace *source_namespace)) *RestFulAuthorNamespace {
	res := &RestFulAuthorNamespace{
		SrcNamespace: &source_namespace{
			//ConditionDefault: &EmptyCondition{},
			//ConditionUpdate:  &EmptyCondition{},
			//ConditionWrite:   &EmptyCondition{},
			//ConditionRead:    &EmptyCondition{},
			//ConditionDelete:  &EmptyCondition{},
		},
		Name:          name,
		SubNamespaces: map[string]*RestFulAuthorNamespace{},
		Parent:        parent,
	}

	for _, op := range ops {
		op(res.SrcNamespace)
	}

	return res
}

func NameSpace(parent *RestFulAuthorNamespace, path string) (res []*RestFulAuthorNamespace, err error) {
	var (
		ok      bool
		tmp     *RestFulAuthorNamespace
		subs    []string
		subsLen int
	)

	subs = PathParse(path)
	if subsLen = len(subs); subsLen == 0 {
		return nil, errors.New("the path is invalid")
	}

	if parent.Name != subs[0] {
		return nil, fmt.Errorf("the first path is not matched")
	} else {
		res = append(res, parent)
		subs = subs[1:]
	}

	for _, sub := range subs {
		if sub == "" {
			continue
		}

		if tmp, ok = parent.SubNamespaces[sub]; ok {
			parent = tmp
			res = append(res, parent)
		} else {
			// * 表示匹配所有
			if tmp, ok = parent.SubNamespaces[ParamAll]; ok {
				parent = tmp
				res = append(res, parent)
			} else {
				return nil, fmt.Errorf("the sub element %s in path %s was not found", sub, path)
			}
		}
	}

	return
}

func AuthorityCheck(namespace *RestFulAuthorNamespace, method string, permission interface{}) (err error) {
	namespace.SrcNamespace.l.Lock()
	defer namespace.SrcNamespace.l.Unlock()

	cond := namespace.SrcNamespace.ConditionDefault
	switch strings.ToUpper(method) {
	case http.MethodGet:
		{
			if nil != namespace.SrcNamespace.ConditionRead {
				cond = namespace.SrcNamespace.ConditionRead
			}
		}
	case http.MethodPost:
		{
			if nil != namespace.SrcNamespace.ConditionWrite {
				cond = namespace.SrcNamespace.ConditionWrite
			}
		}
	case http.MethodDelete:
		{
			if nil != namespace.SrcNamespace.ConditionDelete {
				cond = namespace.SrcNamespace.ConditionDelete
			}
		}
	case http.MethodPut:
		{
			if nil != namespace.SrcNamespace.ConditionUpdate {
				cond = namespace.SrcNamespace.ConditionDefault
			}
		}
	default:

	}

	if cond != nil {
		err = cond.Check(permission)
	}

	return
}
