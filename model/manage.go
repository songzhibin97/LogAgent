package model

import (
	"context"
	"sync"
	"time"

	"github.com/songzhibin97/gkit/goroutine"
)

// AdminMsg 管理结构
type AdminMsg struct {
	sync.RWMutex
	// map[path]*config
	msgMap    map[string]*TailInfo
	kafkaChan chan *LogMsg
	watchChan chan *TailInfo
	Pool      goroutine.GGroup
}

func (a *AdminMsg) GetMsgChan() *LogMsg {
	return <-a.kafkaChan
}

func (a *AdminMsg) PushMsgChan(msg *LogMsg) {
	a.kafkaChan <- msg
}

func (a *AdminMsg) GetWatchChan() *TailInfo {
	return <-a.watchChan
}

func (a *AdminMsg) PushWatchChan(info *TailInfo) {
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

func (a *AdminMsg) AddTask(f func()) bool {
	return a.Pool.AddTask(f)
}

func InitManage(ctx context.Context, maxKafkaChan, maxWatchChan int) *AdminMsg {
	ret := &AdminMsg{
		msgMap:    make(map[string]*TailInfo),
		kafkaChan: make(chan *LogMsg, maxKafkaChan),
		watchChan: make(chan *TailInfo, maxWatchChan),
		Pool:      goroutine.NewGoroutine(ctx, goroutine.SetMax(int64(maxKafkaChan)), goroutine.SetStopTimeout(3*time.Second)),
	}
	ret.goroutineGetWatchChan()
	return ret
}
