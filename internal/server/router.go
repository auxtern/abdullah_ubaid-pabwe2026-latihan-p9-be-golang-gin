package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "app/docs"
	"app/internal/config"
	"app/internal/features/admin"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/todo"
	"app/internal/middleware"
	"app/internal/response"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(1000, time.Second))

	// Repositories & Services
	userRepo := auth.NewUserRepository(db)
	authSvc := auth.NewAuthService(userRepo, cfg.AppKey)
	authHdr := auth.NewAuthHandler(authSvc)

	todoRepo := todo.NewTodoRepository(db)
	todoSvc := todo.NewTodoService(todoRepo)
	todoHdr := todo.NewTodoHandler(todoSvc)

	searchClient := ai.NewSearchEngineClient(cfg.SearchEngineEndpoint)
	aiClient := ai.NewOpenAiClient(cfg.OpenAIEndpoint, cfg.OpenAIKey, cfg.OpenAIModel)
	aiSvc := ai.NewAiService(aiClient, searchClient)
	aiHdr := ai.NewAiHandler(aiSvc)

	adminSvc := admin.NewAdminService(db)
	adminHdr := admin.NewAdminHandler(adminSvc)

	// Swagger Documentation Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		// Auth Routes (Public)
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHdr.Register)
			authRoutes.POST("/login", authHdr.Login)
			authRoutes.POST("/logout", authHdr.Logout)

			// Protected Profile Routes
			authProtected := authRoutes.Group("")
			authProtected.Use(middleware.Auth(cfg.AppKey))
			{
				authProtected.GET("/profile", authHdr.GetProfile)
				authProtected.PUT("/profile", authHdr.UpdateProfile)
			}
		}

		// Public Template Routes
		api.GET("/templates", todoHdr.GetPublicTemplates)
		api.GET("/templates/:id", todoHdr.GetPublicTemplateByID)

		// Protected Routes
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.AppKey))
		{
			// Clone template
			protected.POST("/templates/:id/clone", todoHdr.CloneTemplate)

			// Todos
			protected.GET("/todos", todoHdr.GetTodos)
			protected.POST("/todos", todoHdr.CreateTodo)
			protected.GET("/todos/stats", todoHdr.GetStats)
			protected.GET("/todos/:id", todoHdr.GetTodoByID)
			protected.PUT("/todos/:id", todoHdr.UpdateTodo)
			protected.PATCH("/todos/:id/status", todoHdr.UpdateTodoStatus)
			protected.DELETE("/todos/:id", todoHdr.DeleteTodo)

			// Groups
			protected.GET("/todo-groups", todoHdr.GetGroups)
			protected.POST("/todo-groups", todoHdr.CreateGroup)
			protected.GET("/todo-groups/:id", todoHdr.GetGroupByID)
			protected.PUT("/todo-groups/:id", todoHdr.UpdateGroup)
			protected.DELETE("/todo-groups/:id", todoHdr.DeleteGroup)

			// AI Todo Generator
			protected.POST("/todos/ai-generate", aiHdr.GenerateTodos)

			// Admin Routes
			adm := protected.Group("/admin")
			adm.Use(middleware.RequireRole("admin"))
			{
				adm.GET("/users", adminHdr.GetAllUsers)
				adm.GET("/users/:id/todos", adminHdr.GetUserTodos)
			}
		}
	}

	// Pure REST API 404 handler for unmatched routes
	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, http.StatusNotFound, "Endpoint not found")
	})

	return r
}
