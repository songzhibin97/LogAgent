package Getkafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/olivere/elastic/v7"
	"strings"
	"time"
)

// 定义存储消息结构体
type msgInfo struct {
	Topic    string
	SaveTime *time.Time
	Msg      string
}

var (
	kafkaClient sarama.Consumer // 全局变量
	EsClient    *elastic.Client
	msgChan     chan *msgInfo // 定义管道
)

func Init(address []string, esAddress string, maxChan int) {
	var err error
	config := sarama.NewConfig()
	// 发送完数据需要leader和follow都确认
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 新选出一个partition 模式为随机分配
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// 成功交付的消息将在success channel返回
	config.Producer.Return.Successes = true
	// 连接kafka addr 支持多个地址
	kafkaClient, err = sarama.NewConsumer(address, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}
	fmt.Println("kafka初始完毕")
	// 初始化chan
	msgChan = make(chan *msgInfo, maxChan)
	err = esInit(esAddress)
	if err != nil {
		fmt.Println(err)
	}
	// 开启协程遍历分区并收集消息
	all, err := getAll()
	if err != nil {
		fmt.Println(err)
	}
	for _, topic := range all {
		fmt.Println("启动topic读取kafka--", topic)
		go getMsg(topic)
	}
	// chan中的消息发送到es中
	go kafkaMsgToEs()

}

// 拿到client中所有topic
func getAll() (list []string, err error) {
	topics, err := kafkaClient.Topics()
	if err != nil {
		return nil, err
	}
	for _, topic := range topics {
		if strings.HasPrefix(topic, "_") {
			// 如果前缀是 _ 跳过
			continue
		}
		list = append(list, topic)
	}
	return list, err
}

// 获取消息列表
func getMsg(topic string) (err error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	fmt.Println("进入getMsg")
	partitionList, err := kafkaClient.Partitions(topic) // 根据topic取到所有的分区
	if err != nil {
		fmt.Printf("fail to get list of partition:err%v\n", err)
		return err
	}
	for partition := range partitionList { // 遍历所有的分区
		go func(id int32) {
			// 针对每个分区创建一个对应的分区消费者
			pc, err := kafkaClient.ConsumePartition(topic, id, sarama.OffsetOldest)
			if err != nil {
				fmt.Printf("failed to start consumer for partition %d,err:%v\n", partition, err)
				return
			}
			defer pc.AsyncClose()
			// 异步从每个分区消费信息
			for {
				select {
				case msg := <-pc.Messages():
					fmt.Printf("msg offset: %d, partition: %d, timestamp: %s, value: %s\n",
						msg.Offset, msg.Partition, msg.Timestamp.String(), string(msg.Value))
					msgChan <- &msgInfo{
						Topic:    topic,
						SaveTime: &msg.Timestamp,
						Msg:      string(msg.Value),
					}
				case err := <-pc.Errors():
					fmt.Printf("err :%s\n", err.Error())
				}
			}
		}(int32(partition))
	}
	return
}

func kafkaMsgToEs() {
	if msgChan != nil {
		for {
			select {
			case msg := <-msgChan:
				esSendMsg(msg)
			default:
				time.Sleep(time.Second)
			}

		}
	}
}

func esInit(address string) (err error) {
	EsClient, err = elastic.NewClient(elastic.SetURL(address))
	if err != nil {
		return err
	}
	fmt.Println("es初始完毕")
	return nil
}
func esSendMsg(msg *msgInfo) error {

	put1, err := EsClient.Index().Index(msg.Topic).BodyJson(msg).Do(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("Indexed user %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
	return nil
}
