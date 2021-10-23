package model

import (
	"context"

	"github.com/hpcloud/tail"
)

// PathInfo 配置详情 key:Topic value:Path
type PathInfo struct {
	Topic string `json:"topic"`
	// Path  需要收集的绝对路径
	Path string `json:"path"`
}

// LogMsg 发送至kafka 搜集的队列信息
type LogMsg struct {
	Topic   string
	Message *tail.Line
}

// Kv 用于map存储的临时结构体
type Kv struct {
	Key   string
	Value *TailInfo
}

// TailInfo tail的载体
type TailInfo struct {
	PathInfo
	ctx    context.Context
	cancel context.CancelFunc
}

func (t *TailInfo) Cancel() {
	t.cancel()
}

func (t *TailInfo) GetCtx() context.Context {
	return t.ctx
}

func (t *TailInfo) SetCtx(ctx context.Context) {
	t.ctx = ctx
}

func (t *TailInfo) SetCancel(cancel context.CancelFunc) {
	t.cancel = cancel
}

func CreateTailInfo(ctx context.Context, info PathInfo) *TailInfo {
	_ctx, cancel := context.WithCancel(ctx)
	return &TailInfo{
		PathInfo: info,
		ctx:      _ctx,
		cancel:   cancel,
	}
}
