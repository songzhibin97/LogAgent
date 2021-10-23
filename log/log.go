package tailLog

import (
	"Songzhibin/LogAgent/model"
	"context"
	"fmt"
	"strings"

	"github.com/hpcloud/tail"
)

func InitLog(admin model.IAdminMsg, ctx context.Context, topic string, logPath string) {
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
	fmt.Println("Tail Init Success!", topic)
	go ReadLog(admin, ctx, tails, topic)
}

func ReadLog(admin model.IAdminMsg, ctx context.Context, tails *tail.Tail, topic string) (err error) {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case lines := <-tails.Lines:
			// 处理错误
			if lines.Err != nil {
				return err
			}

			// 处理空文本
			if len(lines.Text) == 0 || len(strings.TrimSpace(lines.Text)) == 0 {
				continue
			}
			admin.PushKafkaChanMsg(&model.LogMsg{
				Topic:   topic,
				Message: lines,
			})
		}
	}
}
