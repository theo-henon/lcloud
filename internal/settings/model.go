package settings

type InstanceSettings struct {
	ID            uint `gorm:"primaryKey"`
	MaskDiskNames bool `gorm:"not null;default:false"`
}

func (InstanceSettings) TableName() string {
	return "instance_settings"
}

type PublicSettings struct {
	MaskDiskNames bool `json:"mask_disk_names"`
}

type UpdateSettingsInput struct {
	MaskDiskNames *bool `json:"mask_disk_names"`
}
