// internal/models/user.go
package models
import (
	"time"
	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	RoleManager UserRole = "manager"
	RoleUser    UserRole = "user"  
	RoleGuest   UserRole = "guest"
)

type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"uniqueIndex;not null"`
	Email          string    `json:"email" gorm:"uniqueIndex;not null"`
	Password       string    `json:"-" gorm:"not null"`
	Role           UserRole  `json:"role" gorm:"not null;default:'user'"`
	NextcloudUser  string    `json:"nextcloud_user"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"` // Para usuarios Guest
	TeamID         *uint     `json:"team_id,omitempty"`
	Team           *Team     `json:"team,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

func (u *User) CanManageTeams() bool {
	return u.Role == RoleManager
}

func (u *User) CanUploadFiles() bool {
	return u.Role == RoleManager || u.Role == RoleUser
}

func (u *User) IsReadOnly() bool {
	return u.Role == RoleGuest
}