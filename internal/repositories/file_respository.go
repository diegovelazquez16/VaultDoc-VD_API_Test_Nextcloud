// internal/repositories/file_repository.go
package repositories

import (
	"test_nextcloud/internal/models"
	"gorm.io/gorm"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (fr *FileRepository) Create(file *models.File) (*models.File, error) {
	if err := fr.db.Create(file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *FileRepository) GetByID(id uint) (*models.File, error) {
	var file models.File
	if err := fr.db.First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (fr *FileRepository) GetByIDWithDetails(id uint) (*models.File, error) {
	var file models.File
	if err := fr.db.Preload("Owner").Preload("Folder").Preload("SharedWith").First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (fr *FileRepository) Update(id uint, updates map[string]interface{}) error {
	return fr.db.Model(&models.File{}).Where("id = ?", id).Updates(updates).Error
}

func (fr *FileRepository) Delete(id uint) error {
	// Eliminar también los permisos asociados
	fr.db.Where("resource_id = ? AND resource = ?", id, "file").Delete(&models.Permission{})
	return fr.db.Delete(&models.File{}, id).Error
}

func (fr *FileRepository) GetUserFiles(userID uint, page, limit int, folderPath, search string) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	query := fr.db.Model(&models.File{}).Where("owner_id = ?", userID)

	if folderPath != "" {
		query = query.Where("path LIKE ?", folderPath+"%")
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Contar total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Obtener archivos paginados
	offset := (page - 1) * limit
	if err := query.Preload("Owner").Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (fr *FileRepository) IsFileSharedWithUser(fileID, userID uint) (bool, error) {
	var count int64
	err := fr.db.Model(&models.Permission{}).
		Where("resource_id = ? AND resource = ? AND user_id = ?", fileID, "file", userID).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

func (fr *FileRepository) GetSharedFiles(userID uint, page, limit int) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	// Subconsulta para obtener IDs de archivos compartidos con el usuario
	subQuery := fr.db.Model(&models.Permission{}).
		Select("resource_id").
		Where("user_id = ? AND resource = ?", userID, "file")

	query := fr.db.Model(&models.File{}).Where("id IN (?)", subQuery)

	// Contar total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Obtener archivos paginados
	offset := (page - 1) * limit
	if err := query.Preload("Owner").Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (fr *FileRepository) CreatePermission(permission *models.Permission) (*models.Permission, error) {
	if err := fr.db.Create(permission).Error; err != nil {
		return nil, err
	}
	return permission, nil
}

func (fr *FileRepository) CreateFolder(folder *models.Folder) (*models.Folder, error) {
	if err := fr.db.Create(folder).Error; err != nil {
		return nil, err
	}
	return folder, nil
}

func (fr *FileRepository) SyncFiles(userID uint, nextcloudFiles []map[string]interface{}) (map[string]interface{}, error) {
	var added, updated, removed int

	tx := fr.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Obtener archivos existentes
	var existingFiles []models.File
	if err := tx.Where("owner_id = ?", userID).Find(&existingFiles).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	existingMap := make(map[string]*models.File)
	for i := range existingFiles {
		existingMap[existingFiles[i].NextcloudID] = &existingFiles[i]
	}

	// Procesar archivos de Nextcloud
	for _, ncFile := range nextcloudFiles {
		ncID := ncFile["id"].(string)
		
		if existingFile, exists := existingMap[ncID]; exists {
			// Actualizar archivo existente si es necesario
			if existingFile.Size != int64(ncFile["size"].(float64)) {
				updates := map[string]interface{}{
					"size":       int64(ncFile["size"].(float64)),
					"updated_at": ncFile["modified"],
				}
				if err := tx.Model(existingFile).Updates(updates).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
				updated++
			}
			delete(existingMap, ncID) // Marcar como procesado
		} else {
			// Crear nuevo archivo
			newFile := &models.File{
				Name:        ncFile["name"].(string),
				Path:        ncFile["path"].(string),
				Size:        int64(ncFile["size"].(float64)),
				MimeType:    ncFile["mime_type"].(string),
				NextcloudID: ncID,
				OwnerID:     userID,
			}
			if err := tx.Create(newFile).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			added++
		}
	}

	// Eliminar archivos que ya no están en Nextcloud
	for _, file := range existingMap {
		if err := tx.Delete(file).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		removed++
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"added":   added,
		"updated": updated,
		"removed": removed,
	}, nil
}