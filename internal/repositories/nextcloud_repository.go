// internal/repositories/nextcloud_repository.go
package repositories

import (
	
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"test_nextcloud/internal/config"
)

type NextcloudRepository struct {
	client *http.Client
	config *config.NextcloudConfig
}

func NewNextcloudRepository(cfg *config.NextcloudConfig) *NextcloudRepository {
	return &NextcloudRepository{
		client: &http.Client{},
		config: cfg,
	}
}

func (nr *NextcloudRepository) UploadFile(username string, file *multipart.FileHeader, folderPath string) (string, error) {
	// Abrir archivo
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Crear request para WebDAV
	url := fmt.Sprintf("%s/remote.php/dav/files/%s%s/%s", 
		nr.config.BaseURL, username, folderPath, file.Filename)

	req, err := http.NewRequest("PUT", url, src)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(username, nr.config.AdminPassword) // O token del usuario
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := nr.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	// Obtener ID del archivo creado
	fileID, err := nr.getFileID(username, folderPath+"/"+file.Filename)
	if err != nil {
		return "", err
	}

	return fileID, nil
}

func (nr *NextcloudRepository) DownloadFile(username, fileID string) ([]byte, error) {
	// Primero obtener la ruta del archivo
	filePath, err := nr.getFilePath(username, fileID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/remote.php/dav/files/%s%s", nr.config.BaseURL, username, filePath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(username, nr.config.AdminPassword)

	resp, err := nr.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (nr *NextcloudRepository) DeleteFile(username, fileID string) error {
	filePath, err := nr.getFilePath(username, fileID)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/remote.php/dav/files/%s%s", nr.config.BaseURL, username, filePath)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, nr.config.AdminPassword)

	resp, err := nr.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (nr *NextcloudRepository) CreateFolder(username, folderPath string) (string, error) {
	url := fmt.Sprintf("%s/remote.php/dav/files/%s%s", nr.config.BaseURL, username, folderPath)

	req, err := http.NewRequest("MKCOL", url, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(username, nr.config.AdminPassword)

	resp, err := nr.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("folder creation failed with status: %d", resp.StatusCode)
	}

	// Obtener ID de la carpeta creada
	folderID, err := nr.getFileID(username, folderPath)
	if err != nil {
		return "", err
	}

	return folderID, nil
}

func (nr *NextcloudRepository) GetUserFiles(username string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/ocs/v2.php/apps/files/api/v1/files", nr.config.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(username, nr.config.AdminPassword)
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Accept", "application/json")

	resp, err := nr.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	files, ok := result["ocs"].(map[string]interface{})["data"].([]interface{})
	if !ok {
		return []map[string]interface{}{}, nil
	}

	var fileList []map[string]interface{}
	for _, file := range files {
		fileMap := file.(map[string]interface{})
		fileList = append(fileList, fileMap)
	}

	return fileList, nil
}

func (nr *NextcloudRepository) getFileID(username, filePath string) (string, error) {
	// Implementar lógica para obtener ID del archivo desde Nextcloud
	// Esto depende de la API específica de Nextcloud
	return "temp_id", nil
}

func (nr *NextcloudRepository) getFilePath(username, fileID string) (string, error) {
	// Implementar lógica para obtener path del archivo desde ID
	// Esto depende de la API específica de Nextcloud
	return "/temp_path", nil
}
