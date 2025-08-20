// internal/services/file_service.go
package services

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"time"

	"test_nextcloud/internal/models"
	"test_nextcloud/internal/repositories"
)

type FileService struct {
	fileRepo      *repositories.FileRepository
	nextcloudRepo *repositories.NextcloudRepository
	userRepo      *repositories.UserRepository
}

func NewFileService(fileRepo *repositories.FileRepository, nextcloudRepo *repositories.NextcloudRepository, userRepo *repositories.UserRepository) *FileService {
	return &FileService{
		fileRepo:      fileRepo,
		nextcloudRepo: nextcloudRepo,
		userRepo:      userRepo,
	}
}

func (fs *FileService) UploadFile(userID uint, file *multipart.FileHeader, folderPath string) (*models.File, error) {
	user, err := fs.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !user.CanUploadFiles() {
		return nil, errors.New("user doesn't have upload permissions")
	}

	// Subir archivo a Nextcloud
	nextcloudFileID, err := fs.nextcloudRepo.UploadFile(user.NextcloudUser, file, folderPath)
	if err != nil {
		return nil, err
	}

	// Crear registro en base de datos
	fileModel := &models.File{
		Name:        file.Filename,
		Path:        filepath.Join(folderPath, file.Filename),
		Size:        file.Size,
		MimeType:    file.Header.Get("Content-Type"),
		NextcloudID: nextcloudFileID,
		OwnerID:     userID,
	}

	createdFile, err := fs.fileRepo.Create(fileModel)
	if err != nil {
		// Si falla el registro en BD, eliminar archivo de Nextcloud
		fs.nextcloudRepo.DeleteFile(user.NextcloudUser, nextcloudFileID)
		return nil, err
	}

	return createdFile, nil
}

func (fs *FileService) DownloadFile(fileID uint) ([]byte, string, error) {
	file, err := fs.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, "", err
	}

	user, err := fs.userRepo.GetByID(file.OwnerID)
	if err != nil {
		return nil, "", err
	}

	fileData, err := fs.nextcloudRepo.DownloadFile(user.NextcloudUser, file.NextcloudID)
	if err != nil {
		return nil, "", err
	}

	return fileData, file.Name, nil
}

func (fs *FileService) GetUserFiles(userID uint, page, limit int, folderPath, search string) ([]models.File, int64, error) {
	return fs.fileRepo.GetUserFiles(userID, page, limit, folderPath, search)
}

func (fs *FileService) HasFileAccess(userID, fileID uint) (bool, error) {
	file, err := fs.fileRepo.GetByID(fileID)
	if err != nil {
		return false, err
	}

	// El propietario siempre tiene acceso
	if file.OwnerID == userID {
		return true, nil
	}

	// Verificar si el archivo ha sido compartido con el usuario
	return fs.fileRepo.IsFileSharedWithUser(fileID, userID)
}

func (fs *FileService) ShareFile(ownerID, fileID uint, userIDs []uint, permissions, expiresAt string) ([]models.Permission, error) {
	// Verificar que el usuario sea el propietario
	file, err := fs.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, err
	}

	if file.OwnerID != ownerID {
		return nil, errors.New("only file owner can share files")
	}

	var expiration *time.Time
	if expiresAt != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04:05Z", expiresAt)
		if err != nil {
			return nil, errors.New("invalid expiration date format")
		}
		expiration = &parsedTime
	}

	var createdPermissions []models.Permission
	for _, userID := range userIDs {
		permission := models.Permission{
			UserID:     userID,
			ResourceID: fileID,
			Resource:   "file",
			Type:       models.PermissionType(permissions),
			ExpiresAt:  expiration,
		}

		createdPerm, err := fs.fileRepo.CreatePermission(&permission)
		if err != nil {
			return nil, err
		}
		createdPermissions = append(createdPermissions, *createdPerm)
	}

	// Marcar archivo como compartido
	fs.fileRepo.Update(fileID, map[string]interface{}{"is_shared": true})

	return createdPermissions, nil
}

func (fs *FileService) DeleteFile(userID, fileID uint) error {
	file, err := fs.fileRepo.GetByID(fileID)
	if err != nil {
		return err
	}

	if file.OwnerID != userID {
		return errors.New("only file owner can delete files")
	}

	user, err := fs.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Eliminar de Nextcloud
	if err := fs.nextcloudRepo.DeleteFile(user.NextcloudUser, file.NextcloudID); err != nil {
		return err
	}

	// Eliminar de base de datos
	return fs.fileRepo.Delete(fileID)
}

func (fs *FileService) CreateFolder(userID uint, name, path string, teamID *uint) (*models.Folder, error) {
	user, err := fs.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(path, name)

	// Crear carpeta en Nextcloud
	nextcloudFolderID, err := fs.nextcloudRepo.CreateFolder(user.NextcloudUser, fullPath)
	if err != nil {
		return nil, err
	}

	folder := &models.Folder{
		Name:        name,
		Path:        fullPath,
		NextcloudID: nextcloudFolderID,
		OwnerID:     userID,
		TeamID:      teamID,
	}

	return fs.fileRepo.CreateFolder(folder)
}

func (fs *FileService) GetSharedFiles(userID uint, page, limit int) ([]models.File, int64, error) {
	return fs.fileRepo.GetSharedFiles(userID, page, limit)
}

func (fs *FileService) GetFileInfo(fileID uint) (*models.File, error) {
	return fs.fileRepo.GetByIDWithDetails(fileID)
}

func (fs *FileService) SyncWithNextcloud(userID uint) (map[string]interface{}, error) {
	user, err := fs.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Obtener archivos de Nextcloud
	nextcloudFiles, err := fs.nextcloudRepo.GetUserFiles(user.NextcloudUser)
	if err != nil {
		return nil, err
	}

	// Sincronizar con base de datos local
	syncResult, err := fs.fileRepo.SyncFiles(userID, nextcloudFiles)
	if err != nil {
		return nil, err
	}

	return syncResult, nil
}
