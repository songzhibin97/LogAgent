package log

import (
	"Songzhibin/LogAgent/model"
	"context"
	"strings"

	"github.com/hpcloud/tail"
)

func ReadLog(admin model.IAdminMsg, ctx context.Context, tails *tail.Tail, topic string) (err error) {

	for {
		select {
		case err = <-ctx.Done():
			return err
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
