package main

import (
	ETcd "Songzhibin/LogAgent/Etcd"
	"Songzhibin/LogAgent/LogMag"
	"Songzhibin/LogAgent/Sendkafka"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"sync"
)

// 读取配置文件反射结构体
type config struct {
	KafkaConfig   *kafkaConfig  `ini:"kafka"`
	ETcd          *etcdConfig   `ini:"etcd"`
	LogMagConfigs *logMagConfig `ini:"manage"`
}
type kafkaConfig struct {
	Address []string
}

type etcdConfig struct {
	Address []string
	Title   string
}

type logMagConfig struct {
	MaxKafkaChan int
	MaxWatchChan int
}

var (
	logPath string
	Configs *config
	wg      sync.RWMutex
	wait  sync.WaitGroup

)

// 初始化
func init() {
	// 读取配置文件
	flag.StringVar(&logPath, "logPath", "/Users/songzhibin/go/src/Songzhibin/LogAgent/serverRun/config.ini", "读取配置文件地址")
	// 解析配置文件
	flag.Parse()
	//
	Configs = new(config)
	err := ini.MapTo(Configs, logPath)
	if err != nil {
		fmt.Println(err)
		panic("读取配置文件失败")
	}
	fmt.Printf("%#v===%#v====%#v\n", Configs.KafkaConfig, Configs.ETcd, Configs.LogMagConfigs)
}
func main() {
	// etcd初始化
	wait.Add(1)
	go LogMag.Init(Configs.LogMagConfigs.MaxKafkaChan, Configs.LogMagConfigs.MaxWatchChan)
	go Sendkafka.Init(Configs.KafkaConfig.Address)
	go ETcd.Init(Configs.ETcd.Address, Configs.ETcd.Title, &wg)
	wait.Wait()
}
