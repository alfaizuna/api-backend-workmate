package server

import (
	"net/http"
	"strings"
	"time"

	"backend-work-mate/internal/auth"
	"backend-work-mate/internal/config"
	"backend-work-mate/internal/storage/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	r := gin.Default()

	userRepo := postgres.NewUserRepository(pool)
	authSvc := auth.NewService(userRepo, cfg)

	// Healthz godoc
	// @Summary Health check
	// @Tags Misc
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /healthz [get]
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"response_code": http.StatusOK,
			"status":        "ok",
		})
	})

	// Swagger UI with explicit doc.json
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
	))

	// Simple JWT auth middleware
	authMW := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"response_code": http.StatusUnauthorized, "error": "missing bearer token"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := auth.ParseAndValidateJWT(tokenString, []byte(cfg.JWTSecret))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"response_code": http.StatusUnauthorized, "error": err.Error()})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}

	api := r.Group("/api")
	{
		// Register godoc
		// @Summary Register user baru
		// @Tags Auth
		// @Accept json
		// @Produce json
		// @Param request body auth.RegisterInput true "Register payload"
		// @Success 201 {object} map[string]interface{}
		// @Failure 400 {object} map[string]interface{}
		// @Router /api/register [post]
		api.POST("/register", func(c *gin.Context) {
			var in auth.RegisterInput
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			user, err := authSvc.Register(c.Request.Context(), in)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"response_code": http.StatusCreated,
				"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role, "department": user.Department, "created_at": user.CreatedAt,
			})
		})

		// Login godoc
		// @Summary Login user
		// @Tags Auth
		// @Accept json
		// @Produce json
		// @Param request body auth.LoginInput true "Login payload"
		// @Success 200 {object} map[string]interface{}
		// @Failure 400 {object} map[string]interface{}
		// @Failure 401 {object} map[string]interface{}
		// @Router /api/login [post]
		api.POST("/login", func(c *gin.Context) {
			var in auth.LoginInput
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			token, err := authSvc.Login(c.Request.Context(), in)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"response_code": http.StatusUnauthorized, "error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "token": token.Token})
		})
	}

	// Tasks routes (protected)
	tasksRepo := postgres.NewTaskRepository(pool)
	tasks := r.Group("/api/tasks", authMW)
	{
		// Create Task
		// @Summary Buat task
		// @Tags Tasks
		// @Security BearerAuth
		// @Accept json
		// @Produce json
		// @Param request body struct{Title string `json:"title"`; Description *string `json:"description"`; Status *string `json:"status"`; DueDate *string `json:"due_date"`} true "Task payload"
		// @Success 201 {object} map[string]interface{}
		// @Failure 400 {object} map[string]interface{}
		// @Router /api/tasks [post]
		tasks.POST("", func(c *gin.Context) {
			var in struct {
				Title       string  `json:"title" binding:"required"`
				Description *string `json:"description"`
				Status      *string `json:"status"`
				DueDate     *string `json:"due_date"`
			}
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			uid := c.GetString("user_id")
			t := &postgres.Task{UserID: uid, Title: in.Title}
			if in.Description != nil {
				t.Description = in.Description
			}
			if in.Status != nil {
				t.Status = *in.Status
			} else {
				t.Status = "Todo"
			}
			if in.DueDate != nil && *in.DueDate != "" {
				if dt, err := time.Parse(time.RFC3339, *in.DueDate); err == nil {
					t.DueDate = &dt
				}
			}
			if err := tasksRepo.Create(c.Request.Context(), t); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"response_code": http.StatusCreated, "data": t})
		})

		// List Tasks
		// @Summary List task milik user
		// @Tags Tasks
		// @Security BearerAuth
		// @Produce json
		// @Success 200 {object} map[string]interface{}
		// @Router /api/tasks [get]
		tasks.GET("", func(c *gin.Context) {
			uid := c.GetString("user_id")
			items, err := tasksRepo.ListByUser(c.Request.Context(), uid, 50, 0)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "data": items})
		})

		// Get Task by ID
		// @Summary Detail task
		// @Tags Tasks
		// @Security BearerAuth
		// @Produce json
		// @Param id path string true "Task ID"
		// @Success 200 {object} map[string]interface{}
		// @Failure 404 {object} map[string]interface{}
		// @Router /api/tasks/{id} [get]
		tasks.GET(":id", func(c *gin.Context) {
			uid := c.GetString("user_id")
			id := c.Param("id")
			t, err := tasksRepo.GetByID(c.Request.Context(), uid, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"response_code": http.StatusNotFound, "error": "not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "data": t})
		})

		// Update Task
		// @Summary Update task
		// @Tags Tasks
		// @Security BearerAuth
		// @Accept json
		// @Produce json
		// @Param id path string true "Task ID"
		// @Param request body struct{Title *string `json:"title"`; Description *string `json:"description"`; Status *string `json:"status"`; DueDate *string `json:"due_date"`} true "Task update"
		// @Success 200 {object} map[string]interface{}
		// @Failure 400 {object} map[string]interface{}
		// @Failure 404 {object} map[string]interface{}
		// @Router /api/tasks/{id} [put]
		tasks.PUT(":id", func(c *gin.Context) {
			var in struct {
				Title       *string `json:"title"`
				Description *string `json:"description"`
				Status      *string `json:"status"`
				DueDate     *string `json:"due_date"`
			}
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			uid := c.GetString("user_id")
			id := c.Param("id")
			t, err := tasksRepo.GetByID(c.Request.Context(), uid, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"response_code": http.StatusNotFound, "error": "not found"})
				return
			}
			if in.Title != nil {
				t.Title = *in.Title
			}
			if in.Description != nil {
				t.Description = in.Description
			}
			if in.Status != nil {
				t.Status = *in.Status
			}
			if in.DueDate != nil {
				if *in.DueDate == "" {
					t.DueDate = nil
				} else if dt, err := time.Parse(time.RFC3339, *in.DueDate); err == nil {
					t.DueDate = &dt
				}
			}
			if err := tasksRepo.Update(c.Request.Context(), t); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "data": t})
		})

		// Delete Task
		// @Summary Hapus task
		// @Tags Tasks
		// @Security BearerAuth
		// @Produce json
		// @Param id path string true "Task ID"
		// @Success 200 {object} map[string]interface{}
		// @Router /api/tasks/{id} [delete]
		tasks.DELETE(":id", func(c *gin.Context) {
			uid := c.GetString("user_id")
			id := c.Param("id")
			if err := tasksRepo.Delete(c.Request.Context(), uid, id); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "message": "deleted"})
		})
	}

	return r
}
