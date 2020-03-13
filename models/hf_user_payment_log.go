package models

import (
	"database/sql"
)

type HfUserPaymentLog struct {
	Id         string `gorm:"primary_key"`
	UserId     string
	OrderId    sql.NullString
	PackageId  string
	IsFirst    uint8
}

func (HfUserPaymentLog) TableName() string {
	return "hf_user_payment_log"
}
