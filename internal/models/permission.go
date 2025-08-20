// internal/models/permission.go
package models

import "time"

type PermissionType string

const (
	PermissionRead     PermissionType = "read"
	PermissionWrite    PermissionType = "write"
	PermissionDelete   PermissionType = "delete"
	PermissionShare    PermissionType = "share"
	PermissionAdmin    PermissionType = "admin"
)

type Permission struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id"`
	User       User           `json:"user" gorm:"foreignKey:UserID"`
	ResourceID uint           `json:"resource_id"`
	Resource   string         `json:"resource"` // "file" or "folder"
	Type       PermissionType `json:"type"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type ActivityLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Action    string    `json:"action"`    // "upload", "download", "share", "delete", etc.
	Resource  string    `json:"resource"`  // archivo o carpeta afectada
	Details   string    `json:"details"`   // información adicional
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}