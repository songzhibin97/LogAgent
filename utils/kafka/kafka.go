package kafka

import (
	"Songzhibin/LogAgent/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/olivere/elastic/v7"

	"github.com/Shopify/sarama"
)

func InitKafka(address []string) (client sarama.Consumer, err error) {
	config := sarama.NewConfig()
	// 发送完数据需要leader和follow都确认
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 新选出一个partition 模式为随机分配
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// 成功交付的消息将在success channel返回
	config.Producer.Return.Successes = true
	// 连接kafka addr 支持多个地址
	client, err = sarama.NewConsumer(address, config)
	if err != nil {
		return
	}
	return client, err
}

// Sentinel 哨兵
func Sentinel(kafkaClient sarama.Consumer, esClient *elastic.Client, channel chan *model.KafkaMsgInfo) {
	go func() {
		all, err := GetAllTopic(kafkaClient)
		if err != nil {
			return
		}
		for _, topic := range all {
			go GetMsgList(kafkaClient, topic, channel)
		}
	}()
	go func() {
		MsgToEs(esClient, channel)
	}()
}

// GetAllTopic 获取所有topic
func GetAllTopic(client sarama.Consumer) (list []string, err error) {
	topics, err := client.Topics()
	for _, topic := range topics {
		if strings.HasPrefix(topic, "_") {
			// 如果是 _ 开头的 跳过
			continue
		}
		list = append(list, topic)
	}
	return
}

// GetMsgList 获取消息列表
func GetMsgList(client sarama.Consumer, topic string, channel chan *model.KafkaMsgInfo) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	partitionList, err := client.Partitions(topic) // 根据topic取到所有的分区
	if err != nil {
		return err
	}
	// 遍历所有的分区
	for id := range partitionList {
		go func(id int) {
			pc, err := client.ConsumePartition(topic, int32(id), sarama.OffsetOldest)
			if err != nil {
				fmt.Printf("failed to start consumer for partition %d,err:%v\n", id, err)
				return
			}
			defer pc.AsyncClose()
			// 异步从每个分区消费信息
			for {
				select {
				case msg := <-pc.Messages():
					if msg == nil {
						return
					}
					channel <- &model.KafkaMsgInfo{
						Topic:    topic,
						Msg:      string(msg.Value),
						SaveTime: &msg.Timestamp,
					}
				case err := <-pc.Errors():
					fmt.Printf("Consumption Kafka Err: %v\n", err)
					return
				}
			}
		}(id)
	}
	return err
}

func MsgToEs(client *elastic.Client, channel chan *model.KafkaMsgInfo) error {
	if channel == nil {
		return errors.New("client is not nil")
	}
	for info := range channel {
		if err := esSendMsg(client, info); err != nil {
			fmt.Printf("MsgToEs err %v\n", err)
		}
	}
	return nil
}

func esSendMsg(client *elastic.Client, msg *model.KafkaMsgInfo) error {
	// todo
	_, err := client.Index().Index(msg.Topic).BodyJson(msg).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// SendMsg 封装发送消息方法
func SendMsg(client sarama.SyncProducer, topic string, message string) error {
	msg := new(sarama.ProducerMessage)
	msg.Topic = topic
	msg.Value = sarama.StringEncoder(message)

	// 发送消息
	_, _, err := client.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

// sendLog 消费消息
func sendLog(client sarama.SyncProducer, admin model.IAdminMsg) {
	go func() {
		defer client.Close()
		for {
			v := admin.GetKafkaChanMsg()
			if v == nil {
				return
			}
			if err := SendMsg(client, v.Topic, v.Message.Text); err != nil {
				// todo log
			}
		}
	}()
}
