package TailLog

import (
	"Songzhibin/LogAgent/LogMag"
	"context"
	"fmt"
	"github.com/hpcloud/tail"
	"time"
)

// TailLog用于读取日志内容
// 使用第三方库 tail


// 初始化模块 更改配置项启动多个init进行多端日志读取
func Init(topic, logPath string, ctx context.Context) () {
	var (
		// tails句柄
		tails *tail.Tail
		err   error
	)

	config := tail.Config{
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: 0,
		},
		Poll:      true,
		ReOpen:    true,
		MustExist: false,
		Follow:    true,
	}
	tails, err = tail.TailFile(logPath, config)
	if err != nil {
		fmt.Println("Create Tail Handle Init error:->", err)
		return
	}
	fmt.Println("Tail Init Success!",topic)
	go readLog(ctx, tails, topic)

}

// 读取日志 建议起一个goroutine后台监听
func readLog(ctx context.Context, tails *tail.Tail, topic string) () {
	for {
		// select多路复用 如果可以从tails.lines取出 *tail.Line 则将其发送至LogChane
		// 否则该goroutine让出cpu进行休眠10s
		select {
		case lines := <-tails.Lines:
			if len(lines.Text) == 0 {
				continue
			}
			fmt.Println("产生了lines:->", lines.Text)
			// lines 需要存放到管道里的消息
			LogMag.ManageMsg.PushKafkaChanMsg(LogMag.NewLog(topic, lines))
		case <-ctx.Done():
			return
		default:
			time.Sleep(2 * time.Second)
		}
	}
}
