package settings

type InstanceSettings struct {
	ID                            uint `gorm:"primaryKey"`
	MaskDiskNames                 bool `gorm:"not null;default:false"`
	ProtocolsWebDAVEnabled        bool `gorm:"not null;default:false"`
	ProtocolsFTPEnabled           bool `gorm:"not null;default:false"`
	NotifyAdminsOnDeletionRequest bool `gorm:"not null;default:true"`
}

func (InstanceSettings) TableName() string {
	return "instance_settings"
}

type PublicSettings struct {
	MaskDiskNames                 bool  `json:"mask_disk_names"`
	ProtocolsWebDAVEnabled        bool  `json:"protocols_webdav_enabled"`
	ProtocolsFTPEnabled           bool  `json:"protocols_ftp_enabled"`
	NotifyAdminsOnDeletionRequest bool  `json:"notify_admins_on_deletion_request"`
	MaxUploadBytes                int64 `json:"max_upload_bytes"`
}

type UpdateSettingsInput struct {
	MaskDiskNames                 *bool `json:"mask_disk_names"`
	ProtocolsWebDAVEnabled        *bool `json:"protocols_webdav_enabled"`
	ProtocolsFTPEnabled           *bool `json:"protocols_ftp_enabled"`
	NotifyAdminsOnDeletionRequest *bool `json:"notify_admins_on_deletion_request"`
}
