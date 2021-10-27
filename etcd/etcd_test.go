package etcd

import (
	"Songzhibin/LogAgent/boot"
	"Songzhibin/LogAgent/local"
	"Songzhibin/LogAgent/model"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitEtcdClient(t *testing.T) {
	// 初始化
	boot.Initialize("../config.yaml")

	client, err := ConnectETCD(local.Config.Etcd.Address)
	if err != nil {
		t.Error(err)
		return
	}
	// 清空配置
	SetValue(client, local.Config.Etcd.Title, "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	local.ManageMsg = model.InitManage(ctx, local.Config.System.MaxKafkaBuffer, local.Config.System.MaxWatchBuffer)
	local.EtcdClient, err = InitEtcdClient(ctx, local.Config.Etcd.Address, local.Config.Etcd.Title)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Len(t, local.ManageMsg.AllKV(), 0)
	SetValue(client, local.Config.Etcd.Title, "[{\"topic\":\"mysql\",\"path\":\"/Users/songzhibin/go/src/Songzhibin/LogAgent/logs/log\"}]")

	time.Sleep(10 * time.Second)
	//assert.Len(t, local.ManageMsg.AllKV(), 1)
	SetValue(client, local.Config.Etcd.Title, "")

	time.Sleep(10 * time.Second)
	//assert.Len(t, local.ManageMsg.AllKV(), 0)
	SetValue(client, local.Config.Etcd.Title, "[{\"topic\":\"mysql\",\"path\":\"/Users/songzhibin/go/src/Songzhibin/LogAgent/logs/log\"},{\"topic\":\"es\",\"path\":\"/Users/songzhibin/go/src/Songzhibin/LogAgent/logs/log\"}]")

	time.Sleep(10 * time.Second)
	//assert.Len(t, local.ManageMsg.AllKV(), 2)
	SetValue(client, local.Config.Etcd.Title, "[{\"topic\":\"mysql\",\"path\":\"/Users/songzhibin/go/src/Songzhibin/LogAgent/logs/log\"}]")

	time.Sleep(10 * time.Second)
	//assert.Len(t, local.ManageMsg.AllKV(), 1)
	SetValue(client, local.Config.Etcd.Title, "")

	time.Sleep(10 * time.Second)
	assert.Len(t, local.ManageMsg.AllKV(), 0)

}
