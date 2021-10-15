package local

import (
	"Songzhibin/LogAgent/model"
	"sync"

	elastic "github.com/olivere/elastic/v7"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/Shopify/sarama"
)

var (
	Lock             sync.RWMutex // 保护下面的全局变量
	EtcdPathInfoList []*model.PathInfo
)

var (
	ManageMsg model.AdminMsg
)

// client

var (
	KafkaClient sarama.Consumer
	EsClient    *elastic.Client
	EtcdClient  *clientv3.Client
)

// channel

var (
	KafkaMsgChannel chan *model.KafkaMsgInfo
)
