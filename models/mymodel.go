package models

type MyModel struct {
	UserId   uint64 `gorm:"primary_key"`
	ServerId uint   `gorm:"primary_key"`
}
