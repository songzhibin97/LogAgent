package tailLog

import (
	"Songzhibin/LogAgent/boot"
	"Songzhibin/LogAgent/local"
	"Songzhibin/LogAgent/model"
	"context"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitLog(t *testing.T) {
	// 初始化
	boot.Initialize("../config.yaml")
	path := "/Users/songzhibin/go/src/Songzhibin/LogAgent/logs/log"
	ctx, cancel := context.WithCancel(context.Background())
	local.ManageMsg = model.InitManage(ctx, local.Config.System.MaxKafkaBuffer, local.Config.System.MaxWatchBuffer)

	InitLog(local.ManageMsg, ctx, "text", path)
	ans := int64(10000)
	go writeLog(cancel, path, ans)

	time.Sleep(time.Second)
	local.ManageMsg.Close()
	for local.ManageMsg.GetMsgChan() != nil {
		atomic.AddInt64(&ans, -1)
	}
	assert.Equal(t, ans, int64(0))

}

func writeLog(cancel context.CancelFunc, path string, n int64) {
	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	defer cancel()

	// 写字节到文件中
	for i := int64(0); i < n; i++ {
		byteSlice := []byte("Bytes!\n")
		_, err := file.Write(byteSlice)
		if err != nil {
			log.Fatal(err)
		}
	}
}
