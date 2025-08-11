package server

import (
	"net/http"

	"backend-work-mate/internal/auth"
	"backend-work-mate/internal/config"
	"backend-work-mate/internal/storage/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	r := gin.Default()

	userRepo := postgres.NewUserRepository(pool)
	authSvc := auth.NewService(userRepo, cfg)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"response_code": http.StatusOK,
			"status":        "ok",
		})
	})

	api := r.Group("/api")
	{
		api.POST("/register", func(c *gin.Context) {
			var in auth.RegisterInput
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"response_code": http.StatusBadRequest,
					"error":         err.Error(),
				})
				return
			}
			user, err := authSvc.Register(c.Request.Context(), in)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"response_code": http.StatusBadRequest,
					"error":         err.Error(),
				})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"response_code": http.StatusCreated,
				"id":            user.ID,
				"name":          user.Name,
				"email":         user.Email,
				"role":          user.Role,
				"department":    user.Department,
				"created_at":    user.CreatedAt,
			})
		})

		api.POST("/login", func(c *gin.Context) {
			var in auth.LoginInput
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"response_code": http.StatusBadRequest,
					"error":         err.Error(),
				})
				return
			}
			token, err := authSvc.Login(c.Request.Context(), in)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"response_code": http.StatusUnauthorized,
					"error":         err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"response_code": http.StatusOK,
				"token":         token.Token,
			})
		})
	}

	return r
}
