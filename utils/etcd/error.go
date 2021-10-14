package etcd

import "github.com/pkg/errors"

var (
	ClientInvalid = errors.New("Etcd客户端无效")
)
