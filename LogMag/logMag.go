package LogMag

import (
	"context"
	"fmt"
	"github.com/hpcloud/tail"
	"net"
	"os"
	"sync"
)

// 提供一些开放给sendKafka 以及 TailLog的一些开放属性

// 定义发送至kafka 收集的对列信息
type logMsg struct {
	Topic   string
	Message *tail.Line
}

// 定义循环map临时结构体
type Kv struct {
	Key   interface{}
	Value interface{}
}

// 定义etcd的任务对列
type EtcdObj struct {
	Topic string `json:"topic"`
	Path  string `json:"path"`
}

// 管理结构
type adminMsg struct {
	// map[path]*config
	aMsg sync.Map
	// KafkaChan: 提供一个暴露在外可以使用的chan 用于异步调用
	kafkaChan chan *logMsg
	// 定义 etcd watch 监控的 topic 提供的chan 用于异步调用
	watchChan chan *Config
}

// 定义接口类
type localAdminMsg interface {
	GetKafkaChanMsg() *logMsg           // 获取kafkaChan数据
	PushKafkaChanMsg(*logMsg)           // 发送kafkaChan数据
	GetWatchChanConfig() *Config        // 获取watchChan数据
	PushWatchChanConfig(*Config)        // 发送watchChan数据
	AddMapKV(string, *Config)           // 增加aMsg中的 sync.Map key:value
	DelMapKV(key string)                // 删除aMsg中的 sync.Map key:value
	ChangeKV(string, *Config) bool      // 改变aMsg中的 sync.Map key:value
	CheckKV(string) (interface{}, bool) // 查找aMsg中的 sync.Map key:value
	AllKV() (sliceKv []*Kv)
	goroutineGetWatchChan()
}

// 定义全局变量
var (
	ManageMsg   localAdminMsg
	EtcdObjList []*EtcdObj
	// markChan 用于初始化记录其他init已经初始化进行后续操作 主要用于etcd控制
	MarkChan chan int
)

// 通过 etcd 获取到的tail配置结构体信息
type Config struct {
	Topic  string
	Path   string
	Ctx    context.Context
	Cancel context.CancelFunc
}

// 提供一些开放的方法
// 构造Msg方法

// Init 初始化管理方法 起协程
func Init(maxKafkaChan, maxWatchChan int) {
	// 读取配置文件 获取最大配置项
	// maxKafkaChan: kafkaChan管道最大容量
	// maxWatchChan: watchChan管道最大容量

	ManageMsg = &adminMsg{
		kafkaChan: make(chan *logMsg, maxKafkaChan),
		watchChan: make(chan *Config, maxWatchChan),
	}
	MarkChan = make(chan int,2)
	// 通道传值 进行确认初始化完成
	fmt.Println("logMag初始化完成")
	MarkChan <- 1
	go ManageMsg.goroutineGetWatchChan()

}

// msg构造方法 (msg -> config)
func NewMsg(topic string, path string, ctx context.Context, cancel context.CancelFunc) (newMsg *Config) {
	return &Config{
		Topic:  topic,
		Path:   path,
		Ctx:    ctx,
		Cancel: cancel,
	}
}

// logMsg构造方法
func NewLog(topic string, messAge *tail.Line) *logMsg {
	return &logMsg{
		Topic:   topic,
		Message: messAge,
	}
}

// 提供 AdminMsg 一些对外的方法

// 获取kafkaChan数据
func (a *adminMsg) GetKafkaChanMsg() *logMsg {
	return <-a.kafkaChan
}

// 发送kafkaChan数据
func (a *adminMsg) PushKafkaChanMsg(msg *logMsg) {
	a.kafkaChan <- msg
}

// 获取watchChan数据
func (a *adminMsg) GetWatchChanConfig() *Config {
	res := <-a.watchChan
	return res
}

// 发送watchChan数据
func (a *adminMsg) PushWatchChanConfig(cof *Config) {
	fmt.Println("塞进", cof)
	a.watchChan <- cof
}

// 维护adminMsg中的 sync.Map
// 增
func (a *adminMsg) AddMapKV(key string, value *Config) {
	a.aMsg.Store(key, value)
	fmt.Println("增加map")
}

// 删
func (a *adminMsg) DelMapKV(key string) {
	a.aMsg.Delete(key)
}

// 改
func (a *adminMsg) ChangeKV(key string, value *Config) bool {
	_, ok := a.aMsg.Load(key)
	if !ok {
		// 如果不存在则返回修改失败
		return false
	}
	a.aMsg.Store(key, value)
	return true
}

// 查
func (a *adminMsg) CheckKV(key string) (value interface{}, isCheck bool) {
	value, isCheck = a.aMsg.Load(key)
	fmt.Println("查到的value,是否存在", value, "====", isCheck)
	return
}

// 获取adminMsg下的aMsg 的所有key:value
func (a *adminMsg) AllKV() (sliceKv []*Kv) {
	fmt.Println("调用range:", a.aMsg)
	sliceKv = make([]*Kv, 0)
	a.aMsg.Range(func(key, value interface{}) bool {
		newKv := &Kv{
			Key:   key,
			Value: value,
		}
		fmt.Println("key", key)
		fmt.Println("value", value)
		fmt.Println(newKv)
		sliceKv = append(sliceKv, newKv)
		return true
	})
	fmt.Printf("sliceKv===%#v", sliceKv)
	return sliceKv
}

// init启动后从watchChan通道获取数据进行异步插入到adminMsg
func (a *adminMsg) goroutineGetWatchChan() {
	fmt.Println("进入监控watchan")
	defer fmt.Println("监控结束")
	for {
		add := a.GetWatchChanConfig()
		a.AddMapKV(add.Topic, add)
	}

}

// 解析获取本机IP
func (a *adminMsg) LocalIp() (ip string, err error) {
	address, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, address := range address {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
			return "", fmt.Errorf("error")
		}
		return "", fmt.Errorf("error")
	}
	return "", fmt.Errorf("error")
}
