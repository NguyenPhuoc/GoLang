package models

type HfUserTowerLog struct {
	UserId string
	Floor  int
	IsWin  int8
	Data   string
}

func (HfUserTowerLog) TableName() string {
	return "hf_user_tower_log"
}
