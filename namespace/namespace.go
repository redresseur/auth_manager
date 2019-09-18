package namespace

import (
	"errors"
	"path/filepath"
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
	name string

	// 子空間
	sub_namespaces map[string]*source_namespace

	// 讀條件集合，每一個資源請求對應的條件
	condition_read Condition

	// 写条件集合
	conditions_write Condition

	// 删除条件集合
	conditions_delete Condition

	// 更新条件集合
	conditions_update Condition

	// 默认条件集合
	conditions_default Condition

	l sync.Mutex
}

type RestFulAuthorNamespace struct {
	srcNamespace *source_namespace
}

func WithReadCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.condition_read = condition
	}
}
func WithWriteCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.conditions_write = condition
	}
}

func WithUpdateCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.conditions_update = condition
	}
}

func WithDeleteCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.conditions_delete = condition
	}
}

func WithDefaultCond(condition Condition) func(namespace *source_namespace) {
	return func(namespace *source_namespace) {
		namespace.conditions_default = condition
	}
}

func UpdateCondition(namespace *RestFulAuthorNamespace, ops ...func(namespace *source_namespace)) error {
	for _, op := range ops {
		op(namespace.srcNamespace)
	}
	return nil
}

// path: 資源路徑
func AddSubNameSpace(parent *RestFulAuthorNamespace, path string) (*RestFulAuthorNamespace, error) {
	var tmp, pNamespace *source_namespace

	sub := &RestFulAuthorNamespace{}
	subs := filepath.SplitList(path)
	if len(subs) == 0 {
		return nil, errors.New("the path is invalid")
	}

	pNamespace = parent.srcNamespace
	for _, s := range subs {
		tmp = &source_namespace{name: s, sub_namespaces: map[string]*source_namespace{}}
		pNamespace.sub_namespaces[s] = tmp
		pNamespace = tmp
	}

	return sub, nil
}

func NewNameSpace(name string, ops ...func(namespace *source_namespace)) *RestFulAuthorNamespace {
	res := &RestFulAuthorNamespace{
		srcNamespace: &source_namespace{
			name:           name,
			sub_namespaces: map[string]*source_namespace{},
		},
	}

	for _, op := range ops {
		op(res.srcNamespace)
	}

	return res
}
