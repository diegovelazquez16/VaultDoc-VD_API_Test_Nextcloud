// internal/routes/routes.go
package routes

import (
	"test_nextcloud/internal/config"
	"test_nextcloud/internal/controllers"
	"test_nextcloud/internal/middleware"
	"test_nextcloud/internal/repositories"
	"test_nextcloud/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Middleware globales
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	// Repositorios
	userRepo := repositories.NewUserRepository(db)
	fileRepo := repositories.NewFileRepository(db)
	nextcloudRepo := repositories.NewNextcloudRepository(&cfg.Nextcloud)

	// Servicios
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	fileService := services.NewFileService(fileRepo, nextcloudRepo, userRepo)

	// Controladores
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	fileController := controllers.NewFileController(fileService)

	// Rutas públicas
	public := router.Group("/api/v1")
	{
		public.POST("/auth/login", authController.Login)
		public.POST("/auth/register", authController.Register)
		
		// Health check
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
				"service": "nextcloud-bridge-api",
			})
		})
	}

	// Rutas protegidas
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuth())
	{
		// Autenticación
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authController.Logout)
			auth.POST("/refresh", authController.RefreshToken)
		}

		// Usuarios
		users := protected.Group("/users")
		{
			users.GET("/profile", userController.GetProfile)
			users.PUT("/profile", userController.UpdateProfile)
			users.GET("/activity", userController.GetActivityReport)
		}

		// Archivos
		files := protected.Group("/files")
		{
			files.POST("/upload", fileController.UploadFile)
			files.GET("/", fileController.GetFiles)
			files.GET("/shared", fileController.GetSharedFiles)
			files.GET("/:id", fileController.GetFileInfo)
			files.GET("/:id/download", fileController.DownloadFile)
			files.POST("/:id/share", fileController.ShareFile)
			files.DELETE("/:id", fileController.DeleteFile)
			files.POST("/sync", fileController.SyncFiles)
		}

		// Carpetas
		folders := protected.Group("/folders")
		{
			folders.POST("/", fileController.CreateFolder)
		}
	}

	// Rutas solo para managers
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.JWTAuth())
	admin.Use(middleware.RequireRole("manager"))
	{
		// Gestión de usuarios
		admin.GET("/users", userController.GetUsers)
		admin.POST("/users", userController.CreateUser)
		admin.DELETE("/users/:id", userController.DeleteUser)
		
		// Reportes
		admin.GET("/reports/activity", userController.GetActivityReport)
	}
}
