package local

import (
	"Songzhibin/LogAgent/configs"
	"Songzhibin/LogAgent/model"
	"sync"

	"github.com/spf13/viper"

	elastic "github.com/olivere/elastic/v7"
	v3 "go.etcd.io/etcd/client/v3"

	"github.com/Shopify/sarama"
)

var (
	Viper *viper.Viper
)

var (
	Lock             sync.RWMutex // 保护下面的全局变量
	EtcdPathInfoList []*model.PathInfo
)

var Config configs.Server

var (
	ManageMsg *model.AdminMsg
)

// client

var (
	KafkaClient    sarama.Consumer
	KafkaCProducer sarama.SyncProducer
	EsClient       *elastic.Client
	EtcdClient     *v3.Client
)

// channel

var (
	KafkaMsgChannel chan *model.KafkaMsgInfo
)
