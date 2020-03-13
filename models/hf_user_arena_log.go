package models

type HfUserArenaLog struct {
	UserAtt string
	UserDef string
	EloAtt      int
	EloDef      int
	EloAttAfter int
	EloDefAfter int
	Data        string
}

func (HfUserArenaLog) TableName() string {
	return "hf_user_arena_log"
}
