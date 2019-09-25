package namespace

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
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
	res := rc.FindAllString(str, 1)
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

//func CopyCondition(dst *RestFulAuthorNamespace, src *RestFulAuthorNamespace )  {
//	dst.SrcNamespace
//}

func ReverseFind(sp *RestFulAuthorNamespace, name string) (*RestFulAuthorNamespace, error) {
	if sp.Name == name {
		return sp, nil
	}

	if sp.Parent == nil {
		return nil, errors.New("Namespace Not Found")
	}

	return ReverseFind(sp.Parent, name)
}

// path: 資源路徑
func AddSubNameSpace(parent *RestFulAuthorNamespace, path string) (*RestFulAuthorNamespace, error) {
	var tmp *RestFulAuthorNamespace

	subs := strings.Split(path, "/")
	if len(subs) == 0 {
		return nil, errors.New("the path is invalid")
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
		tmp = &RestFulAuthorNamespace{
			SrcNamespace: &source_namespace{
				Variable:  isParam,
				ParamName: param,
			},
			Name:          sub,
			Parent:        parent,
			SubNamespaces: map[string]*RestFulAuthorNamespace{},
		}

		parent.SubNamespaces[sub] = tmp
		parent = tmp
	}

	return parent, nil
}

func NewNameSpace(name string, ops ...func(namespace *source_namespace)) *RestFulAuthorNamespace {
	res := &RestFulAuthorNamespace{
		SrcNamespace:  &source_namespace{},
		Name:          name,
		SubNamespaces: map[string]*RestFulAuthorNamespace{},
	}

	for _, op := range ops {
		op(res.SrcNamespace)
	}

	return res
}

func NameSpace(parent *RestFulAuthorNamespace, path string) (res []*RestFulAuthorNamespace, err error) {
	subs := strings.Split(path, "/")
	if len(subs) == 0 {
		return nil, errors.New("the path is invalid")
	}

	var ok bool
	for _, sub := range subs {
		if sub == "" {
			continue
		}

		res = append(res, parent)
		if parent, ok = parent.SubNamespaces[sub]; !ok {
			return nil, fmt.Errorf("the sub element %s in path %s was not found", sub, path)
		}
	}

	return
}

func AuthorityCheck(namespace *RestFulAuthorNamespace, permission interface{}) error {
	namespace.SrcNamespace.l.Lock()
	defer namespace.SrcNamespace.l.Unlock()

	return namespace.SrcNamespace.ConditionDefault.Check(permission)
}
