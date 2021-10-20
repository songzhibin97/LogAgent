//go:build !windows
// +build !windows

package internal

import (
	"Songzhibin/LogAgent/local"
	"os"
	"path"
	"time"

	logs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
)

func GetWriteSyncer() (zapcore.WriteSyncer, error) {
	fileWriter, err := logs.New(
		path.Join(local.Config.Zap.Director, "%Y-%m-%d.log"),
		logs.WithLinkName(local.Config.Zap.LinkName),
		logs.WithMaxAge(7*24*time.Hour),
		logs.WithRotationTime(24*time.Hour),
	)
	if local.Config.Zap.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	}
	return zapcore.AddSync(fileWriter), err
}
