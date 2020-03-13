package models

type HfUserArenaPvpLog struct {
	UserId string
	Elo      int
	IsWin    int
	EloAfter int
	Data     string
}

func (HfUserArenaPvpLog) TableName() string {
	return "hf_user_arena_pvp_log"
}
