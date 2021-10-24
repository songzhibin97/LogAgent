package etcd

import (
	"Songzhibin/LogAgent/local"
	"Songzhibin/LogAgent/model"
	tailLog "Songzhibin/LogAgent/tailLog"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"

	v3 "go.etcd.io/etcd/client/v3"
)

// InitEtcdClient 初始化etcdClient并启动哨兵
func InitEtcdClient(ctx context.Context, address []string, title string) (client *v3.Client, err error) {
	client, err = v3.New(v3.Config{
		Endpoints:   address,          // 连接ip地址 localhost:2379
		DialTimeout: 20 * time.Second, // 超时时间
	})
	if err != nil {
		return nil, err
	}
	childCtx, cancel := context.WithCancel(ctx)
	// 先尝试获取
	resp, err := GetByEtcd(client, title)
	if err != nil {
		panic(err)
	}
	err = decodeConfig(resp, &local.EtcdPathInfoList, &local.Lock)
	if err != nil {
		fmt.Println("decode err", err)
	}
	local.Lock.Lock()
	for _, info := range local.EtcdPathInfoList {
		fmt.Println("info:", info)
		_ctx, _cancel := context.WithCancel(childCtx)
		pushModel := &model.TailInfo{
			PathInfo: *info,
		}
		pushModel.SetCtx(_ctx)
		pushModel.SetCancel(_cancel)
		local.ManageMsg.PushWatchChan(pushModel)
		tailLog.InitLog(local.ManageMsg, _ctx, info.Topic, info.Path)
	}
	local.Lock.Unlock()
	// 启动哨兵监控这个key的变化
	go Sentinel(client, title, childCtx)
	// 后台起协程监控父ctx退出
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		}
	}()
	return client, nil
}

// title : ip/title:[{"topic":"xxx","path":"xxx/xx/xx"},{"topic":"xxx","path":"xxx/xx/xx"}]

// GetByEtcd 从etcd客户端获取key的value
func GetByEtcd(client *v3.Client, key string) (*v3.GetResponse, error) {
	if client == nil {
		return nil, ClientInvalid
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := client.Get(ctx, key)
	return resp, err
}

// PutByEtcd 从etcd客户端设置key的value
func PutByEtcd(client *v3.Client, key string, values string) error {
	if client == nil {
		return ClientInvalid
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err := client.Put(ctx, key, values)
	return err
}

// DeleteByEtcd 从etcd客户端删除key
func DeleteByEtcd(client *v3.Client, key string) error {
	if client == nil {
		return ClientInvalid
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err := client.Delete(ctx, key)
	return err
}

// Sentinel 哨兵
func Sentinel(client *v3.Client, key string, ctx context.Context) {
	// 在初始化init调用的主程不进行保存
	eventChannel := client.Watch(ctx, key)
	for event := range eventChannel {
		for _, _event := range event.Events {

			switch _event.Type {
			case mvccpb.PUT:
				configList := local.ManageMsg.AllKV()
				if len(_event.Kv.Value) == 0 {
					// 循环  logMag/adminMsg/aMsg -> []map{}
					// configList -> []*kv
					// 条件分支: 如果传来的配置文件为0 清除本地所有配置并cancel掉
					for _, kv := range configList {
						kv.Value.Cancel()
						local.ManageMsg.DelMapKV(kv.Value.Topic)
					}
				} else {
					// 临时map
					temporary := make(map[string]*model.Kv, len(configList))
					for _, kv := range configList {
						temporary[kv.Value.Topic] = kv
					}

					// 解析配置
					var temporaryList []*model.PathInfo
					if err := json.Unmarshal(_event.Kv.Value, &temporaryList); err != nil {
						// 解析错误不进行处理
						continue
					}
					// 找差集
					for _, info := range temporaryList {
						if info.Topic == "" || info.Path == "" {
							continue
						}
						_info, ok := local.ManageMsg.CheckKV(info.Topic)
						if !ok {
							// 原来不存在
							local.ManageMsg.PushWatchChan(model.CreateTailInfo(ctx, *info))
							tailLog.InitLog(local.ManageMsg, ctx, info.Topic, info.Path)
						} else {
							// 删除 temporary _info.Topic
							delete(temporary, info.Topic)
							// 判断 path 是否一致
							if ok && (info.Path == _info.Path) {
								continue
							}
							// 两边path不一致,需要删除对象,重新创建
							_info.Cancel()
							local.ManageMsg.DelMapKV(info.Topic)
							local.ManageMsg.AddMapKV(info.Topic, model.CreateTailInfo(ctx, *info))
							tailLog.InitLog(local.ManageMsg, ctx, info.Topic, info.Path)
						}
					}
					for _, kv := range temporary {
						// 多余的删除掉
						kv.Value.Cancel()
						local.ManageMsg.DelMapKV(kv.Value.Topic)
					}
				}
			case mvccpb.DELETE:
				configList := local.ManageMsg.AllKV()
				for _, kv := range configList {
					// 停止服务并且删除local map
					kv.Value.Cancel()
					local.ManageMsg.DelMapKV(kv.Value.Topic)
				}
			}
		}
	}
}

// decodeConfig 解析配置信息
// obj 为decode位置 必须为指针
func decodeConfig(resp *v3.GetResponse, obj interface{}, wg *sync.RWMutex) error {
	for _, kv := range resp.Kvs {
		// 主进程只可能循环一次
		// 解析到全局变量 EtcdObjList(不需要初始化)
		if kv.Value != nil {
			if err := func() error {
				wg.Lock()
				defer wg.Unlock()
				if err := json.Unmarshal(kv.Value, obj); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return err
			}
		}
	}
	return nil
}
