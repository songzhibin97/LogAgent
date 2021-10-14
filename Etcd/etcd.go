package Edct

import (
	"Songzhibin/LogAgent/LogMag"
	"Songzhibin/LogAgent/TailLog"
	"context"
	"encoding/json"
	"fmt"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"sync"
	"time"
)

var (
	ETcdClient *clientv3.Client
)

// etcd初始化方法 起 goroutine
func Init(address []string, title string, wg *sync.RWMutex) {
	var err error
	// 初始化变量 后续初始化条件
	i := 0
	// 创建链接
	ETcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   address,         // 连接ip地址 localhost:2379
		DialTimeout: 5 * time.Second, // 超时时间
	})
	if err != nil {
		panic(err)
	}
	// 调用get方法获取配置
	resp, err := EtcdGet(title)
	if err != nil {
		panic(err)
	}
	// 将config进行解包
	err = decodeConfig(resp, wg)
	if err != nil {
		fmt.Println("err", err)
	}
	for {
		i += <-LogMag.MarkChan
		if i == 2 {
			close(LogMag.MarkChan)
			break
		}
	}
	// 进行logMag的数据初始化
	for _, obj := range LogMag.EtcdObjList {
		fmt.Println("obj", obj)
		// adminMsg.aMsg map增加topic:*config
		ctx, cancel := context.WithCancel(context.Background())
		// 添加到 logMag/adminMsg/watchChan
		LogMag.ManageMsg.PushWatchChanConfig(LogMag.NewMsg(obj.Topic, obj.Path, ctx, cancel))
		// 启动读取日志init
		go TailLog.Init(obj.Topic, obj.Path, ctx)
	}
	// 总哨兵激活
	go EtcdWatcher(title)
}

// Get
func EtcdGet(title string) (*clientv3.GetResponse, error) {
	// title : ip/title:[{"topic":"xxx","path":"xxx/xx/xx"},{"topic":"xxx","path":"xxx/xx/xx"}]
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := ETcdClient.Get(ctx, title)
	cancel()
	return resp, err
}

// Put
func EtcdPut(title string, jsonString string) error {
	// title : ip/title:[{"topic":"xxx","path":"xxx/xx/xx"},{"topic":"xxx","path":"xxx/xx/xx"}]
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := ETcdClient.Put(ctx, title, jsonString)
	cancel()
	return err
}

// Delete
func EtcdDelete(title string) error {
	// title : ip/title:[{"topic":"xxx","path":"xxx/xx/xx"},{"topic":"xxx","path":"xxx/xx/xx"}]
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := ETcdClient.Delete(ctx, title)
	cancel()
	return err
}

// watch 哨兵监控title
func EtcdWatcher(title string) {
	// 在init调用的主哨兵不进行保存
	fmt.Println("进入哨兵")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()                      // 注册退出
	rch := ETcdClient.Watch(ctx, title) // <-chan WatchResponse
	for resp := range rch {
		for _, ev := range resp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			//哨兵监控如果有改动进行数据整理 PUT
			if ev.Type == mvccpb.PUT {
				fmt.Println("进入put分支")
				// 当前logMag中adminMsg下的aMsg的map对列
				configList := LogMag.ManageMsg.AllKV()
				fmt.Printf("目前的configList map维护的键值对========>%#v\n\n", configList)
				if len(ev.Kv.Value) == 0 {
					fmt.Println("进入值为0区域")

					// 循环  logMag/adminMsg/aMsg -> []map{}
					// configList -> []*kv
					// 条件分支: 如果传来的配置文件为0
					for _, kvObj := range configList {
						v, ok := kvObj.Value.(*LogMag.Config)
						if !ok {
							continue
						}
						fmt.Println("注销所有在线线程")
						// 注销所有在线线程
						fmt.Println("调用关闭了", v.Topic)
						v.Cancel()
						// 删除logMag/aMsg 的key:value
						LogMag.ManageMsg.DelMapKV(v.Topic)
					}
				} else {
					fmt.Println("不为0分支")
					// 做一个map映射
					temporaryMap := make(map[string]*LogMag.Config, len(configList))
					for _, kvObj2 := range configList {
						v, ok := kvObj2.Value.(*LogMag.Config)
						if !ok {
							continue
						}
						temporaryMap[v.Topic] = v
					}
					var temporaryList []*LogMag.EtcdObj
					err := json.Unmarshal(ev.Kv.Value, &temporaryList)
					if err != nil {
						fmt.Println("err", err)
						fmt.Println("配置错误 不处理")
						// 如果传来的配置信息有误 不符合条件则不进行处理
						continue
					}
					fmt.Printf("temporaryList 新解析的decode==============>%#v", temporaryList)
					// 条件分支: 如果新传来的配置文件不为0
					for _, kvObj := range temporaryList {
						// 循环新创建的中间变量
						value, ok := LogMag.ManageMsg.CheckKV(kvObj.Topic)
						if !ok {
							fmt.Println("不存在 进行添加", kvObj.Topic)
							// 如果不存在 则进行添加
							// 添加到 logMag/adminMsg/watchChan
							newCtx, newCancel := context.WithCancel(context.Background())
							LogMag.ManageMsg.PushWatchChanConfig(LogMag.NewMsg(kvObj.Topic, kvObj.Path, newCtx, newCancel))
							go TailLog.Init(kvObj.Topic, kvObj.Path, newCtx)
							continue
						} else {
							fmt.Println("存在 准备进行下一步校验")
							//存在 扣除map中对应映射
							fmt.Println("删除了map映射", kvObj.Topic)
							delete(temporaryMap, kvObj.Topic)
							fmt.Printf("还剩下map映射%#v\n", temporaryMap)
							// 如果存在进行下一步校验
							v, ok := value.(*LogMag.Config)
							if !ok {
								// 如果不是*LogMag.Config类型则进行跳过处理
								continue
							}
							// 校验失败 原路径与新路径不符 进行修改
							if v.Path != kvObj.Path {
								fmt.Println("存在 校验失败")
								// 直接删除旧的 新建新的
								LogMag.ManageMsg.DelMapKV(kvObj.Topic)
								// 将原来的删除 重新生成tail对象
								fmt.Println("调用关闭了", v.Topic)
								v.Cancel()

								newCtx, newCancel := context.WithCancel(context.Background())
								// map塞新值
								LogMag.ManageMsg.AddMapKV(v.Topic, LogMag.NewMsg(kvObj.Topic, kvObj.Path, newCtx, newCancel))

								// 重新生成tail对象
								go TailLog.Init(kvObj.Topic, kvObj.Path, newCtx)
								continue
							} else {
								fmt.Println("存在 校验成功")
								continue
							}
						}
					}
					// 循环map映射 将本来有这次没有的删除
					for _, v := range temporaryMap {
						fmt.Println("将原先有这次没有的删除", v.Topic)
						fmt.Println("调用关闭了", v.Topic)
						v.Cancel()
					}
				}
			}
			if ev.Type == mvccpb.DELETE {
				fmt.Println("进入delete分支")
				// 删除处理 全部停止
				// 循环  logMag/adminMsg/aMsg -> []map{}
				// configList -> []*kv
				// 当前logMag中adminMsg下的aMsg的map对列
				configList := LogMag.ManageMsg.AllKV()
				// 条件分支: 如果传来的配置文件为0
				for _, kvObk := range configList {
					v, ok := kvObk.Value.(*LogMag.Config)
					if !ok {
						continue
					}
					// 注销所有在线线程
					fmt.Println("调用关闭了", v.Topic)
					v.Cancel()
					// 删除logMag/aMsg 的key:value
					LogMag.ManageMsg.DelMapKV(v.Topic)
				}
			}
		}
	}
}

// 解析配置信息
func decodeConfig(resp *clientv3.GetResponse, wg *sync.RWMutex) error {
	for _, kv := range resp.Kvs {
		// 主进程只可能循环一次
		// 解析到全局变量 EtcdObjList(不需要初始化)
		if kv.Value != nil {
			wg.Lock()
			defer wg.Unlock()
			err := json.Unmarshal(kv.Value, &LogMag.EtcdObjList)
			if err != nil {
				fmt.Println("err", err)
				return err
			}
		}
	}
	return nil
}
