package main

import (
	"Songzhibin/LogAgent/boot"
	"Songzhibin/LogAgent/etcd"
	"Songzhibin/LogAgent/kafka"
	"Songzhibin/LogAgent/local"
	"Songzhibin/LogAgent/model"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化配置
	boot.Initialize()

	ctx, cancel := context.WithCancel(context.Background())
	local.ManageMsg = model.InitManage(ctx, local.Config.System.MaxKafkaBuffer, local.Config.System.MaxWatchBuffer)
	var err error
	local.KafkaCProducer, err = kafka.InitKafkaProducer(local.Config.Kafka.Address)
	if err != nil {
		panic(err)
	}
	fmt.Println("kafka初始化完成:", local.Config.Kafka.Address)
	kafka.SendLog(local.KafkaCProducer, local.ManageMsg)

	local.EtcdClient, err = etcd.InitEtcdClient(ctx, local.Config.Etcd.Address, local.Config.Etcd.Title)
	if err != nil {
		panic(err)
	}
	fmt.Println("etcd初始化完成")

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
		cancel()
	}
}
