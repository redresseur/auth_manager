package namespace

// TODO: 添加到一个可共享的存储地址
type Manager interface {
	Load(path string) (*RestFulAuthorNamespace, error)
	Store(path string, namespace *RestFulAuthorNamespace) error
}

type EmptyManager struct {
}

func (*EmptyManager) Load(path string) (*RestFulAuthorNamespace, error) {
	return nil, nil
}

func (*EmptyManager) Store(path string, namespace *RestFulAuthorNamespace) error {
	return nil
}

func NewEmptyManager() Manager {
	return &EmptyManager{}
}
