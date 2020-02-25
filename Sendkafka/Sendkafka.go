package Sendkafka

import (
	"Songzhibin/LogAgent/LogMag"
	"fmt"
	"github.com/Shopify/sarama"
)

var (
	kafkaClient sarama.SyncProducer
)

// 初始化模块
func Init(address []string) {
	var err error
	config := sarama.NewConfig()
	// 发送完数据需要leader和follow都确认
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 新选出一个partition 模式为随机分配
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// 成功交付的消息将在success channel返回
	config.Producer.Return.Successes = true
	// 连接kafka addr 支持多个地址
	kafkaClient, err = sarama.NewSyncProducer(address, config) //
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}
	fmt.Println("KafkaClient Init Success!")
	// 向通道传值 确认初始化成功
	LogMag.MarkChan <- 1
	// 起协程进行管道数据收集
	go sendLog()

}

// 封装发送消息方法
func sendMsg(topic string, message string) (err error) {
	msg := new(sarama.ProducerMessage)
	msg.Topic = topic
	msg.Value = sarama.StringEncoder(message)
	// 发送消息
	pid, offset, err := kafkaClient.SendMessage(msg)
	if err != nil {
		fmt.Println("send msg failed, err:", err)
		return
	}
	fmt.Printf("topic:%v,pid:%v offset:%v,value:%v\n", topic, pid, offset, message)
	return nil
}

// 接受chan 建议起goroutine后台轮询发送信息
func sendLog() {
	fmt.Println("进入sedLog")
	defer kafkaClient.Close()
	for {
		ls := LogMag.ManageMsg.GetKafkaChanMsg()
		err := sendMsg(ls.Topic, ls.Message.Text)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
