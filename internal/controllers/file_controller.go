// internal/controllers/file_controller.go
package controllers

import (
	"net/http"
	"strconv"
	

	"github.com/gin-gonic/gin"
	"test_nextcloud/internal/models"
	"test_nextcloud/internal/services"
	"test_nextcloud/pkg/utils"
)

type FileController struct {
	fileService *services.FileService
}

func NewFileController(fileService *services.FileService) *FileController {
	return &FileController{
		fileService: fileService,
	}
}

func (fc *FileController) UploadFile(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	// Verificar permisos de subida
	if !currentUser.CanUploadFiles() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Read-only access")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file provided", err.Error())
		return
	}

	folderPath := c.PostForm("folder_path")
	if folderPath == "" {
		folderPath = "/"
	}

	uploadedFile, err := fc.fileService.UploadFile(currentUser.ID, file, folderPath)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Upload failed", err.Error())
		return
	}

	utils.SuccessResponse(c, uploadedFile)
}

func (fc *FileController) DownloadFile(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	// Verificar permisos de acceso
	hasAccess, err := fc.fileService.HasFileAccess(currentUser.ID, uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Access check failed", err.Error())
		return
	}

	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "No permission to access this file")
		return
	}

	fileData, fileName, err := fc.fileService.DownloadFile(uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "File not found", err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/octet-stream", fileData)
}

func (fc *FileController) GetFiles(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	folderPath := c.Query("folder")
	search := c.Query("search")

	files, total, err := fc.fileService.GetUserFiles(currentUser.ID, page, limit, folderPath, search)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get files", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"files": files,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (fc *FileController) ShareFile(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	if currentUser.IsReadOnly() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Read-only access")
		return
	}

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	var shareRequest struct {
		UserIDs     []uint `json:"user_ids"`
		Permissions string `json:"permissions"` // "read" or "write"
		ExpiresAt   string `json:"expires_at,omitempty"`
	}

	if err := c.ShouldBindJSON(&shareRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	shares, err := fc.fileService.ShareFile(currentUser.ID, uint(fileID), shareRequest.UserIDs, shareRequest.Permissions, shareRequest.ExpiresAt)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Share failed", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "File shared successfully",
		"shares":  shares,
	})
}

func (fc *FileController) DeleteFile(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	if currentUser.IsReadOnly() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Read-only access")
		return
	}

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	if err := fc.fileService.DeleteFile(currentUser.ID, uint(fileID)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Delete failed", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "File deleted successfully"})
}

func (fc *FileController) CreateFolder(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	if !currentUser.CanUploadFiles() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Read-only access")
		return
	}

	var folderRequest struct {
		Name   string `json:"name" binding:"required"`
		Path   string `json:"path"`
		TeamID *uint  `json:"team_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&folderRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	folder, err := fc.fileService.CreateFolder(currentUser.ID, folderRequest.Name, folderRequest.Path, folderRequest.TeamID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Folder creation failed", err.Error())
		return
	}

	utils.SuccessResponse(c, folder)
}

func (fc *FileController) GetSharedFiles(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sharedFiles, total, err := fc.fileService.GetSharedFiles(currentUser.ID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shared files", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"files": sharedFiles,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (fc *FileController) GetFileInfo(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	hasAccess, err := fc.fileService.HasFileAccess(currentUser.ID, uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Access check failed", err.Error())
		return
	}

	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "No permission to access this file")
		return
	}

	fileInfo, err := fc.fileService.GetFileInfo(uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "File not found", err.Error())
		return
	}

	utils.SuccessResponse(c, fileInfo)
}

func (fc *FileController) SyncFiles(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	if currentUser.IsReadOnly() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Read-only access")
		return
	}

	syncResult, err := fc.fileService.SyncWithNextcloud(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Sync failed", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Sync completed successfully",
		"result":  syncResult,
	})
}