package model

import "sync"

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

func InitManage(maxKafkaChan, maxWatchChan int) *AdminMsg {
	ret := &AdminMsg{
		msgMap:    make(map[string]*TailInfo),
		kafkaChan: make(chan *LogMsg, maxKafkaChan),
		watchChan: make(chan *TailInfo, maxWatchChan),
	}
	ret.goroutineGetWatchChan()
	return ret
}
