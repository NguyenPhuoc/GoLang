package models

type HfUserStone struct {
	UserId   string `gorm:"primary_key"`
	Type     string  `gorm:"primary_key"`
	Quantity uint
}

func (HfUserStone) TableName() string {
	return "hf_user_stone"
}

func (HfUserStone) GetTypeConfig() []string {
	return []string{"d", "l", "w", "t", "e", "f", "evo"}
}
