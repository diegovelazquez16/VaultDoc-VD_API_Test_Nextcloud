// internal/controllers/auth_controller.go
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"test_nextcloud/internal/models"
	"test_nextcloud/internal/services"
	"test_nextcloud/pkg/utils"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string          `json:"username" binding:"required"`
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=6"`
	Role     models.UserRole `json:"role"`
	TeamID   *uint           `json:"team_id,omitempty"`
}

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, token, err := ac.authService.Login(req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"user":         user,
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600, // 1 hour
	})
}

func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, err := ac.authService.Register(req.Username, req.Email, req.Password, req.Role, req.TeamID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Registration failed", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"user":    user,
		"message": "User registered successfully",
	})
}

func (ac *AuthController) Logout(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	err := ac.authService.Logout(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Logout failed", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Logged out successfully"})
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	user, newToken, err := ac.authService.RefreshToken(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Token refresh failed", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"user":         user,
		"access_token": newToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}