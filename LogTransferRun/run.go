/******
** @创建时间 : 2020/2/24 09:43
** @作者 : SongZhiBin
******/
package main

import (
	"Songzhibin/LogAgent/Getkafka"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"sync"
)

// 读取配置文件反射结构体
type config struct {
	KafkaConfig   *kafkaConfig  `ini:"kafka"`
	LogMagConfigs *logMagConfig `ini:"manage"`
	ESConfig *esConfig `ini:"es"`
}
type kafkaConfig struct {
	Address []string
}
type esConfig struct {
	Address string
}

type logMagConfig struct {
	MaxKafkaChan int
	MaxWatchChan int
}

var (
	logPath string
	Configs *config
	wait    sync.WaitGroup
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
	fmt.Printf("%#v==%#v\n", Configs.KafkaConfig, Configs.LogMagConfigs)
}

func main() {
	wait.Add(1)
	Getkafka.Init(Configs.KafkaConfig.Address, Configs.ESConfig.Address,Configs.LogMagConfigs.MaxKafkaChan)
	wait.Wait()
}
