package model

import (
	"context"
	"sync"

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

// AdminMsg 管理结构
type AdminMsg struct {
	sync.RWMutex
	// map[path]*config
	msgMap    map[string]*TailInfo
	kafkaChan chan *LogMsg
	watchChan chan *TailInfo
}

func (a *AdminMsg) GetKafkaChanMsg() *LogMsg {
	return <-a.kafkaChan
}

func (a *AdminMsg) PushKafkaChanMsg(msg *LogMsg) {
	a.kafkaChan <- msg
}

func (a *AdminMsg) GetWatchChanConfig() *TailInfo {
	return <-a.watchChan
}

func (a *AdminMsg) PushWatchChanConfig(info *TailInfo) {
	a.watchChan <- info
}

func (a *AdminMsg) AddMapKV(s string, info *TailInfo) {
	a.Lock()
	defer a.Unlock()
	if _, ok := a.msgMap[s]; ok {
		return
	}
	a.msgMap[s] = info
}

func (a *AdminMsg) DelMapKV(key string) {
	a.Lock()
	defer a.Unlock()
	delete(a.msgMap, key)
}

func (a *AdminMsg) ChangeKV(s string, info *TailInfo) bool {
	a.Lock()
	defer a.Unlock()
	_, ok := a.msgMap[s]
	if !ok {
		return false
	}
	a.msgMap[s] = info
	return true
}

func (a *AdminMsg) CheckKV(s string) (*TailInfo, bool) {
	a.Lock()
	defer a.Unlock()
	v, ok := a.msgMap[s]
	return v, ok
}

func (a *AdminMsg) AllKV() (sliceKv []*Kv) {
	a.Lock()
	defer a.Unlock()
	sliceKv = make([]*Kv, 0, len(a.msgMap))
	for key, value := range a.msgMap {
		sliceKv = append(sliceKv, &Kv{
			Key:   key,
			Value: value,
		})
	}
	return sliceKv
}

func (a *AdminMsg) goroutineGetWatchChan() {
	go func() {
		for info := range a.watchChan {
			a.AddMapKV(info.Topic, info)
		}
	}()
}

func CreateTailInfo(ctx context.Context, info PathInfo) *TailInfo {
	_ctx, cancel := context.WithCancel(ctx)
	return &TailInfo{
		PathInfo: info,
		ctx:      _ctx,
		cancel:   cancel,
	}
}
