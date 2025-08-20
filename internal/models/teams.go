// internal/models/team.go
package models

import "time"

type Team struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	ManagerID   uint      `json:"manager_id"`
	Manager     User      `json:"manager" gorm:"foreignKey:ManagerID"`
	Members     []User    `json:"members,omitempty"`
	Folders     []Folder  `json:"folders,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
