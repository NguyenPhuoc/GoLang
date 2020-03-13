package models

type HfUserTokenLog struct {
	UserId string
	Data   string
}

func (HfUserTokenLog) TableName() string {
	return "hf_user_token_log"
}