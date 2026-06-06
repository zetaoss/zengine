package models

type AppSetting struct {
	ID         int    `json:"id" gorm:"column:id"`
	SettingKey string `json:"setting_key" gorm:"column:setting_key"`
	ValueJSON  string `json:"value_json" gorm:"column:value_json"`
}

func (AppSetting) TableName() string {
	return "app_settings"
}
