package boot

import (
	"Songzhibin/LogAgent/boot/internal"
	"Songzhibin/LogAgent/local"
	"Songzhibin/LogAgent/utils"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Zap = new(_zap)

type _zap struct {
	err    error
	level  zapcore.Level
	writer zapcore.WriteSyncer
	zap    *zap.Logger
}

func (z *_zap) Initialize() {
	if ok, _ := utils.PathExists(local.Config.Zap.Director); !ok { // 判断是否有Director文件夹
		fmt.Println(`创建文件夹成功, 文件夹名称为: `, local.Config.Zap.Director)
		_ = os.Mkdir(local.Config.Zap.Director, os.ModePerm)
	}

	switch local.Config.Zap.Level { // 初始化配置文件的Level
	case "debug":
		z.level = zap.DebugLevel
	case "info":
		z.level = zap.InfoLevel
	case "warn":
		z.level = zap.WarnLevel
	case "error":
		z.level = zap.ErrorLevel
	case "dpanic":
		z.level = zap.DPanicLevel
	case "panic":
		z.level = zap.PanicLevel
	case "fatal":
		z.level = zap.FatalLevel
	default:
		z.level = zap.InfoLevel
	}
	if z.writer, z.err = internal.GetWriteSyncer(); z.err != nil { // 使用file-rotatelogs进行日志分割
		fmt.Println(`获取WriteSyncer失败, err: `, z.err)
		return
	}
	//if z.level == zap.DebugLevel || z.level == zap.ErrorLevel {
	//	z.zap = zap.New(internal.GetEncoderCore(z.writer, z.level), zap.AddStacktrace(z.level))
	//} else {
	z.zap = zap.New(internal.GetEncoderCore(z.writer, z.level))
	//}
	if local.Config.Zap.ShowLine {
		z.zap = z.zap.WithOptions(zap.AddCaller())
	}
	zap.ReplaceGlobals(z.zap)
}
