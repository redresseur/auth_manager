package namespace

import (
	"errors"
	"fmt"
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
type source_namespace struct {
	// 空间名称
	Name string `json:"name"`

	// 子空間
	SubNamespaces map[string]*source_namespace `json:"subNamespaces"`

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
	SrcNamespace *source_namespace `json:"srcNamespace"`
}

func WithReadCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.ConditionRead = condition
	}
}
func WithWriteCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.ConditionWrite = condition
	}
}

func WithUpdateCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.ConditionUpdate = condition
	}
}

func WithDeleteCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.ConditionDelete = condition
	}
}

func WithDefaultCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.ConditionDefault = condition
	}
}

func UpdateCondition(namespace *RestFulAuthorNamespace, ops ...func(namespace *source_namespace)) error {
	for _, op := range ops {
		op(namespace.SrcNamespace)
	}
	return nil
}

// path: 資源路徑
func AddSubNameSpace(parent *RestFulAuthorNamespace, path string) (*RestFulAuthorNamespace, error) {
	var tmp, pNamespace *source_namespace

	sub := &RestFulAuthorNamespace{}
	subs := strings.Split(path, "/")
	if len(subs) == 0 {
		return nil, errors.New("the path is invalid")
	}

	pNamespace = parent.SrcNamespace
	for _, s := range subs {
		tmp = &source_namespace{Name: s, SubNamespaces: map[string]*source_namespace{}}
		pNamespace.SubNamespaces[s] = tmp
		pNamespace = tmp
	}

	sub.SrcNamespace = tmp
	return sub, nil
}

func NewNameSpace(name string, ops ...func(namespace *source_namespace)) *RestFulAuthorNamespace {
	res := &RestFulAuthorNamespace{
		SrcNamespace: &source_namespace{
			Name:          name,
			SubNamespaces: map[string]*source_namespace{},
		},
	}

	for _, op := range ops {
		op(res.SrcNamespace)
	}

	return res
}

func NameSpace(parent *RestFulAuthorNamespace, path string) (*RestFulAuthorNamespace, error) {
	subs := strings.Split(path, "/")
	if len(subs) == 0 {
		return nil, errors.New("the path is invalid")
	}

	var ok bool
	var psNamespace *source_namespace = parent.SrcNamespace
	for _, sub := range subs {
		if psNamespace, ok = psNamespace.SubNamespaces[sub]; !ok{
			return nil, fmt.Errorf("the sub element %s in path %s was not found", sub, path)
		}
	}

	return &RestFulAuthorNamespace{
		SrcNamespace: psNamespace,
	}, nil
}
