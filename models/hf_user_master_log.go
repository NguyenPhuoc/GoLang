package models

import (
	"database/sql"
)

type HfUserMasterLog struct {
	UserId        string
	Type          int
	Item          string
	ItemId        string
	KindId        uint
	EventId       int
	Quantity      int
	AfterQuantity uint64
	Note          sql.NullString
}

func (HfUserMasterLog) TableName() string {
	return "hf_user_master_log"
}

func (uml *HfUserMasterLog) SaveLog(u *HfUser, typeLog int, item string, itemId string, kindId uint, eventId int, quantity int, afterQuantity uint64, note sql.NullString) {

	uml.UserId = u.UserId
	uml.Type = typeLog
	uml.Item = item
	uml.ItemId = itemId
	uml.KindId = kindId
	uml.EventId = eventId
	uml.Quantity = quantity
	uml.AfterQuantity = afterQuantity
	uml.Note = note

	//u.DB.Save(&uml)
}
