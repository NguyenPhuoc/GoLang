package models

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"time"
)

type HfUserTransaction struct {
	OrderId        string `gorm:"primary_key"`
	PartnerOrderId sql.NullString
	UserId         string
	ServerId       uint
	PackageId      string
	IsTopup        uint8
	Type           uint8
	TimeTopup      mysql.NullTime
	TimeOrder      time.Time
}

func (HfUserTransaction) TableName() string {
	return "hf_user_transaction"
}
