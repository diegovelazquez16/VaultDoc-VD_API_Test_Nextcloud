// internal/services/auth_service.go
package services

import (
	"errors"
	"time"

	"test_nextcloud/internal/models"
	"test_nextcloud/internal/repositories"
	"test_nextcloud/pkg/utils"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (as *AuthService) Login(username, password string) (*models.User, string, error) {
	user, err := as.userRepo.GetByUsername(username)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, "", errors.New("account is disabled")
	}

	if user.IsExpired() {
		return nil, "", errors.New("account has expired")
	}

	if !user.CheckPassword(password) {
		return nil, "", errors.New("invalid credentials")
	}

	// Generar JWT token
	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return nil, "", err
	}

	// Actualizar último login
	now := time.Now()
	user.LastLoginAt = &now
	as.userRepo.Update(user.ID, map[string]interface{}{"last_login_at": now})

	// Limpiar password del response
	user.Password = ""

	return user, token, nil
}

func (as *AuthService) Register(username, email, password string, role models.UserRole, teamID *uint) (*models.User, error) {
	// Verificar si el usuario ya existe
	if _, err := as.userRepo.GetByUsername(username); err == nil {
		return nil, errors.New("username already exists")
	}

	if _, err := as.userRepo.GetByEmail(email); err == nil {
		return nil, errors.New("email already exists")
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Role:     role,
		TeamID:   teamID,
		IsActive: true,
	}

	// Hash password
	if err := user.HashPassword(password); err != nil {
		return nil, err
	}

	// Para usuarios Guest, establecer expiración por defecto (7 días)
	if role == models.RoleGuest {
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		user.ExpiresAt = &expiresAt
	}

	createdUser, err := as.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	createdUser.Password = ""
	return createdUser, nil
}

func (as *AuthService) Logout(userID uint) error {
	// Aquí podrías invalidar tokens si usas un blacklist
	// Por ahora solo registramos la actividad
	return nil
}

func (as *AuthService) RefreshToken(userID uint) (*models.User, string, error) {
	user, err := as.userRepo.GetByID(userID)
	if err != nil {
		return nil, "", errors.New("user not found")
	}

	if !user.IsActive || user.IsExpired() {
		return nil, "", errors.New("account is not active")
	}

	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return nil, "", err
	}

	user.Password = ""
	return user, token, nil
}
