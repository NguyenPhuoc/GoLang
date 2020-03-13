package models

import (
	"github.com/go-sql-driver/mysql"
	"sync"
)

type HfUserItems struct {
	UserId    string `gorm:"primary_key"`
	Item      string `gorm:"primary_key"`
	Quantity  int
	TimeCheck mysql.NullTime
}

func (HfUserItems) TableName() string {
	return "hf_user_items"
}

func (ui *HfUserItems) Get(material string, u HfUser) (HfUserItems, bool) {
	ui.UserId = u.UserId
	ui.Item = material

	count := 0
	u.DB.Where(ui).First(&ui).Count(&count)

	return *ui, count != 0
}

func (ui *HfUserItems) Set(quantity int, typeLog, eventId int, u HfUser, wg *sync.WaitGroup) {
	defer wg.Done()

	ui.Quantity += quantity

	wg.Add(2)
	go u.SaveLog(typeLog, ui.Item, "", 0, eventId, quantity, uint64(ui.Quantity), "", wg)
	go func() {
		u.DB.Save(&ui)
		wg.Done()
	}()
}
