// internal/services/user_service.go  
package services

import (
	"test_nextcloud/internal/models"
	"test_nextcloud/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (us *UserService) GetByID(userID uint) (*models.User, error) {
	return us.userRepo.GetByID(userID)
}

func (us *UserService) Update(userID uint, updates map[string]interface{}) (*models.User, error) {
	if err := us.userRepo.Update(userID, updates); err != nil {
		return nil, err
	}
	return us.userRepo.GetByID(userID)
}

func (us *UserService) GetPaginated(page, limit int, role string) ([]models.User, int64, error) {
	return us.userRepo.GetPaginated(page, limit, role)
}

func (us *UserService) Create(username, email, password string, role models.UserRole, teamID *uint) (*models.User, error) {
	user := &models.User{
		Username: username,
		Email:    email,
		Role:     role,
		TeamID:   teamID,
		IsActive: true,
	}

	if err := user.HashPassword(password); err != nil {
		return nil, err
	}

	return us.userRepo.Create(user)
}

func (us *UserService) Delete(userID uint) error {
	return us.userRepo.Delete(userID)
}

func (us *UserService) GetActivityReport(userID uint, from, to string) ([]models.ActivityLog, error) {
	return us.userRepo.GetActivityReport(userID, from, to)
}
