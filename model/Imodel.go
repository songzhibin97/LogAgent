package model

import "context"

// ITailInfo TailInfo接口类
type ITailInfo interface {
	Cancel()
	GetCtx() context.Context
	SetCtx(ctx context.Context)
	SetCancel(cancel context.CancelFunc)
}

type IAdminMsg interface {
	GetKafkaChanMsg() *LogMsg         // 获取kafkaChan数据
	PushKafkaChanMsg(*LogMsg)         // 发送kafkaChan数据
	GetWatchChanConfig() *TailInfo    // 获取watchChan数据
	PushWatchChanConfig(*TailInfo)    // 发送watchChan数据
	AddMapKV(string, *TailInfo)       // 增加msgMap中的 Map key:value
	DelMapKV(key string)              // 删除msgMap中的 Map key:value
	ChangeKV(string, *TailInfo) bool  // 改变msgMap中的 Map key:value
	CheckKV(string) (*TailInfo, bool) // 查找msgMap中的 Map key:value
	AllKV() (sliceKv []*Kv)
	goroutineGetWatchChan()
	AddTask(f func()) bool
}
