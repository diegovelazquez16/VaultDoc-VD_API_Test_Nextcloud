// internal/models/file.go
package models

import "time"

type File struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Path         string    `json:"path" gorm:"not null"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	NextcloudID  string    `json:"nextcloud_id"`
	OwnerID      uint      `json:"owner_id"`
	Owner        User      `json:"owner" gorm:"foreignKey:OwnerID"`
	FolderID     *uint     `json:"folder_id,omitempty"`
	Folder       *Folder   `json:"folder,omitempty"`
	IsShared     bool      `json:"is_shared" gorm:"default:false"`
	SharedWith   []User    `json:"shared_with,omitempty" gorm:"many2many:file_shares;"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Folder struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Path        string    `json:"path" gorm:"not null"`
	NextcloudID string    `json:"nextcloud_id"`
	TeamID      *uint     `json:"team_id,omitempty"`
	Team        *Team     `json:"team,omitempty"`
	OwnerID     uint      `json:"owner_id"`
	Owner       User      `json:"owner" gorm:"foreignKey:OwnerID"`
	Files       []File    `json:"files,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
