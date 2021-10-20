package boot

import (
	"Songzhibin/LogAgent/local"
	"flag"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	ConfigEnv  = "Agent"
	ConfigFile = "config.yaml"
)

var Viper = new(_viper)

type _viper struct {
	err  error
	path string
}

func (v *_viper) Initialize(path ...string) {
	if len(path) == 0 {
		flag.StringVar(&v.path, "c", "", "choose config file.")
		flag.Parse()
		if v.path == "" { // 优先级: 命令行 > 环境变量 > 默认值
			if configEnv := os.Getenv(ConfigEnv); configEnv == "" {
				v.path = ConfigEnv
				fmt.Println(`您正在使用环境变量,config的路径为: `, v.path)
			} else {
				fmt.Println(`您正在使用命令行的-c参数传递的值,config的路径为: `, v.path)
			}
		} else {
			v.path = path[0]
			fmt.Println(`您正在使用func Viper()传递的值,config的路径为: `, v.path)
		}
		v.path = ConfigFile
		fmt.Println(`您正在使用config的默认值, config的路径为: `, v.path)
		var _v = viper.New()
		_v.SetConfigFile(v.path)
		if v.err = _v.ReadInConfig(); v.err != nil {
			panic(fmt.Sprintf(`读取config.yaml文件失败, err: %v`, v.err))
		}
		_v.WatchConfig()

		_v.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println(`配置文件已修改并更新,文件为: `, e.Name)
			if v.err = _v.Unmarshal(&local.Config); v.err != nil {
				fmt.Println(v.err)
			}
		})
		if v.err = _v.Unmarshal(&local.Config); v.err != nil {
			fmt.Println(`Json 序列化数据失败, err :`, v.err)
		}
		local.Viper = _v
	}
}
