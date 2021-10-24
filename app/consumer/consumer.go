package main

import (
	"Songzhibin/LogAgent/boot"
	"Songzhibin/LogAgent/es"
	"Songzhibin/LogAgent/kafka"
	"Songzhibin/LogAgent/local"
	"Songzhibin/LogAgent/model"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化配置
	boot.Initialize()

	var err error
	local.KafkaClient, err = kafka.InitKafka(local.Config.Kafka.Address)
	if err != nil {
		panic(err)
	}
	fmt.Println("kafka初始化完成:", local.Config.Kafka.Address)
	// 初始化channel对象
	local.KafkaMsgChannel = make(chan *model.KafkaMsgInfo, local.Config.System.MaxKafkaBuffer)

	// 初始化ES对象
	fmt.Println(local.Config.Es.Address)
	local.EsClient, err = es.InitEs(local.Config.Es.Address)
	if err != nil {
		panic(err)
	}
	// 开启协程遍历分区并接受消息
	all, err := kafka.GetAllTopic(local.KafkaClient)
	if err != nil {
		fmt.Println("GetAllTopic:", err)
	}
	fmt.Println("all topic:", all)
	for _, topic := range all {
		if err = kafka.GetMsgList(local.KafkaClient, topic, local.KafkaMsgChannel); err != nil {
			fmt.Println("GetMsgList err:", err)
		}
	}
	// 开启消费者
	for i := 0; i < local.Config.System.MaxSentinel; i++ {
		go kafka.MsgToEs(local.EsClient, local.KafkaMsgChannel)
	}

	// 监听信号
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	// 阻塞在此，当接收到上述两种信号时才会往下执行
	// 创建一个5秒超时的context

	select {
	case <-quit:

	}
}
