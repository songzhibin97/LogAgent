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
	GetMsgChan() *LogMsg              // 获取Msg数据
	PushMsgChan(*LogMsg)              // 发送Msg数据
	GetWatchChan() *TailInfo          // 获取watchChan数据
	PushWatchChan(*TailInfo)          // 发送watchChan数据
	AddMapKV(string, *TailInfo)       // 增加msgMap中的 Map key:value
	DelMapKV(key string)              // 删除msgMap中的 Map key:value
	ChangeKV(string, *TailInfo) bool  // 改变msgMap中的 Map key:value
	CheckKV(string) (*TailInfo, bool) // 查找msgMap中的 Map key:value
	AllKV() (sliceKv []*Kv)           // 获取所有kv
	goroutineGetWatchChan()           // 后台哨兵
	AddTask(f func()) bool            // 异步任务队列
}
