// pkg/utils/response.go
package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	response := Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusOK, response)
}

func SuccessResponseWithMessage(c *gin.Context, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusOK, response)
}

func ErrorResponse(c *gin.Context, statusCode int, message, error string) {
	response := Response{
		Success:   false,
		Message:   message,
		Error:     error,
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

func ValidationErrorResponse(c *gin.Context, errors map[string]string) {
	response := map[string]interface{}{
		"success":          false,
		"message":          "Validation failed",
		"validation_errors": errors,
		"timestamp":        time.Now(),
	}
	c.JSON(http.StatusBadRequest, response)
}
