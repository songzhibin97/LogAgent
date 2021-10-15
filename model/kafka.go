package model

import "time"

type KafkaMsgInfo struct {
	Topic    string
	Msg      string
	SaveTime *time.Time
}
