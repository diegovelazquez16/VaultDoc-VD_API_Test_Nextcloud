// internal/controllers/user_controller.go
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"test_nextcloud/internal/models"
	"test_nextcloud/internal/services"
	"test_nextcloud/pkg/utils"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	user, err := uc.userService.GetByID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.SuccessResponse(c, user)
}

func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, err := uc.userService.Update(userID, updates)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Update failed", err.Error())
		return
	}

	utils.SuccessResponse(c, user)
}

func (uc *UserController) GetUsers(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	// Solo managers pueden ver todos los usuarios
	if !currentUser.CanManageTeams() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Insufficient permissions")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	role := c.Query("role")

	users, total, err := uc.userService.GetPaginated(page, limit, role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get users", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (uc *UserController) CreateUser(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	if !currentUser.CanManageTeams() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Insufficient permissions")
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, err := uc.userService.Create(req.Username, req.Email, req.Password, req.Role, req.TeamID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "User creation failed", err.Error())
		return
	}

	utils.SuccessResponse(c, user)
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	if !currentUser.CanManageTeams() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Insufficient permissions")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	if err := uc.userService.Delete(uint(userID)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to delete user", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "User deleted successfully"})
}

func (uc *UserController) GetActivityReport(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	if !currentUser.CanManageTeams() {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "Insufficient permissions")
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	from := c.Query("from")
	to := c.Query("to")

	activities, err := uc.userService.GetActivityReport(uint(userID), from, to)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get activity report", err.Error())
		return
	}

	utils.SuccessResponse(c, activities)
}