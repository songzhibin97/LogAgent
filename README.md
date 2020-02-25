### LogAgent && LogTransfer
#### 使用场景
- 需要高集群高并发多端日志收集管理
- 需要动态更改日志项提供热加载
- 灾难处理机制以及特殊的异常处理
- 全平台兼容 Windows/Linux/Mac



#### 目录结构
```
LogAgent // 主目录
// etcd Package 提供了一些操作etcd的封装方法
|_______Etcd	
	|_______etcd.go 
// logMag Package 提供了LogAgent中所有使用的接口、结构体、以及需要的中间变量的存储声明 并开放在外的一些操作内部结构体的方法,提供chan监控消费者分发消息等方法
|_______LogMag 
	|_______logMag.go
// SendKafka Package 提供发送kafka,接受kafka固定topic消息的内容
|_______SendKafka
	|_______Sendkafka.go
// TailLog Package 提供固定topic以及path抓取指定path并且封装成结构内容发送至缓存通道 供消费使用
|_______TailLog
	|_______ReadLog.go
// 读取ini配置文件初始化LogAgent 并发开启多线程初始化logMag、kafka、etcd ini文件可以在shell命令行中指定位置 -logPath=path
|_______serverRun
	|_______run.go
	|_______config.ini
// 新增 logTransfer 用于读取kafka中的日志 发送到ES中
|_______LogTransferRun
	|_______run.go
// 新增 从kafka读取全部topic的内容高并发监控
|_______Getkafka
	|_______Getkafka.go
```
#### 配置文件结构
```
# kafka ip
[kafka]
Address = 192.168.125.102:9092

# etcd ip 以及 title 如果集群建议增加 ip/title
[etcd]
Address = 127.0.0.1:2379
Title = text

# MaxKafkaChan 最大缓存kafka数量
# MaxWatchChan 最大缓存监控配置数量
[manage]
MaxKafkaChan = 10000
MaxWatchChan = 1000

# 新增es配置项
[es]
Address = http://127.0.0.1:9200
```
#### ReadMe
- 下载
```go
go get github.com/songzhibin97/LogAgent
```
- 启动客户端 ``` 在serverRun下 有编译好的各平台客户端 直接运行即可 ```
- 在LogAgent下提供一个发送etcd的test.go 修改当中的 ```Endpoints(ETCD IP)``` ```jsonString(配置Json串)```运行即可发送 
``` jsonString(配置Json串) 格式 ```
```
`[{"topic":"topic","path":"xxx.xx"}]`
```

- 启动客户端 logTransfer 收集kafka的日志并发送到ES中 依赖ES 如果需要可视化需要kibana

