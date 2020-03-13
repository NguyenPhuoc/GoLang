package models

type HfUserSetting struct {
	UserId  string `gorm:"primary_key"`
	Setting string `gorm:"default:'{}'"`
}

func (HfUserSetting) TableName() string {
	return "hf_user_setting"
}

func (m *HfUserSetting) Get(u HfUser) HfUserSetting {
	m.UserId = u.UserId

	count := 0
	u.DB.Where(m).Find(&m).Count(&count)

	if count == 0 {
		u.DB.Save(&m)
	}

	return *m
}
