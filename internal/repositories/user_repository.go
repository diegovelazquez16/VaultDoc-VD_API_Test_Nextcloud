// internal/repositories/user_repository.go
package repositories

import (
	"test_nextcloud/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) Create(user *models.User) (*models.User, error) {
	if err := ur.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := ur.db.Preload("Team").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := ur.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := ur.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) Update(id uint, updates map[string]interface{}) error {
	return ur.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (ur *UserRepository) Delete(id uint) error {
	return ur.db.Delete(&models.User{}, id).Error
}

func (ur *UserRepository) GetPaginated(page, limit int, role string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := ur.db.Model(&models.User{})
	
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Contar total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Obtener usuarios paginados
	offset := (page - 1) * limit
	if err := query.Preload("Team").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (ur *UserRepository) GetActivityReport(userID uint, from, to string) ([]models.ActivityLog, error) {
	var activities []models.ActivityLog
	query := ur.db.Where("user_id = ?", userID)

	if from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to != "" {
		query = query.Where("created_at <= ?", to)
	}

	if err := query.Order("created_at DESC").Find(&activities).Error; err != nil {
		return nil, err
	}

	return activities, nil
}
