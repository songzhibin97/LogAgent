package local

import (
	"Songzhibin/LogAgent/model"
	"sync"
)

var (
	Lock             sync.RWMutex // 保护下面的全局变量
	EtcdPathInfoList []*model.PathInfo
)

var (
	ManageMsg model.AdminMsg
)
