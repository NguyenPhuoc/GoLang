package models

import (
	"GoLang/libraries/util"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
	"time"
)

type HfUserInbox struct {
	Id           string `gorm:"primary_key"`
	ReceiverId   string
	SenderType   uint8 `gorm:"default:1"`//loại gửi từ đâu
	SenderId     sql.NullString
	FullName     sql.NullString
	ServerId     uint
	TypeLog      int
	EventId      int
	KindId       int
	IsRead       uint8
	IsReceive    uint8
	IsDelete     uint8
	Title        string
	Message      sql.NullString
	Gift         sql.NullString
	Note         sql.NullString
	ReadDate     mysql.NullTime
	ReceivedDate mysql.NullTime
	DeletedDate  mysql.NullTime
	CreatedDate  time.Time
}

func (HfUserInbox) TableName() string {
	return "hf_user_inbox"
}

func (ui *HfUserInbox) GetMap() iris.Map {
	return iris.Map{
		"id":           ui.Id,
		"receiver_id":  ui.ReceiverId,
		"sender_type":  ui.SenderType,
		"sender_id":    ui.SenderId.String,
		"full_name":    ui.FullName.String,
		"server_id":    ui.ServerId,
		"kind_id":      ui.KindId,
		"is_read":      ui.IsRead,
		"is_receive":   ui.IsReceive,
		"title":        ui.Title,
		"message":      ui.Message.String,
		"gift":         util.JsonDecodeMap(ui.Gift.String),
		"created_date": util.TimeToDateTime(ui.CreatedDate),
	}
}
