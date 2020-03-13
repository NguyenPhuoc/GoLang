package models

type HfUserEventLog struct {
	UserId  string
	EventId int
	Type    uint
	KindId  uint
	Data    string
}

func (HfUserEventLog) TableName() string {
	return "hf_user_event_log"
}
