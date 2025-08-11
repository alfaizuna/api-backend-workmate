package server

import (
	"net/http"
	"time"

	"backend-work-mate/internal/auth"
	"backend-work-mate/internal/storage/postgres"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	AuthSvc   *auth.Service
	TaskRepo  postgres.TaskRepository
	JWTSecret []byte
}

// Healthz godoc
// @Summary Health check
// @Tags Misc
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /healthz [get]
func (h *Handlers) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "status": "ok"})
}

// Register godoc
// @Summary Register user baru
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterInput true "Register payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/register [post]
func (h *Handlers) Register(c *gin.Context) {
	var in auth.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
		return
	}
	user, err := h.AuthSvc.Register(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
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
}

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
func (h *Handlers) Login(c *gin.Context) {
	var in auth.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
		return
	}
	token, err := h.AuthSvc.Login(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"response_code": http.StatusUnauthorized, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "token": token.Token})
}

type CreateTaskInput struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	DueDate     *string `json:"due_date"`
}

type UpdateTaskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	DueDate     *string `json:"due_date"`
}

// Create Task godoc
// @Summary Buat task
// @Tags Tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateTaskInput true "Task payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/tasks [post]
func (h *Handlers) CreateTask(c *gin.Context) {
	var in CreateTaskInput
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
	if err := h.TaskRepo.Create(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"response_code": http.StatusCreated, "data": t})
}

// List Tasks godoc
// @Summary List task milik user
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/tasks [get]
func (h *Handlers) ListTasks(c *gin.Context) {
	uid := c.GetString("user_id")
	items, err := h.TaskRepo.ListByUser(c.Request.Context(), uid, 50, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "data": items})
}

// Get Task by ID godoc
// @Summary Detail task
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/tasks/{id} [get]
func (h *Handlers) GetTask(c *gin.Context) {
	uid := c.GetString("user_id")
	id := c.Param("id")
	t, err := h.TaskRepo.GetByID(c.Request.Context(), uid, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"response_code": http.StatusNotFound, "error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "data": t})
}

// Update Task godoc
// @Summary Update task
// @Tags Tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body UpdateTaskInput true "Task update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/tasks/{id} [put]
func (h *Handlers) UpdateTask(c *gin.Context) {
	var in UpdateTaskInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
		return
	}
	uid := c.GetString("user_id")
	id := c.Param("id")
	t, err := h.TaskRepo.GetByID(c.Request.Context(), uid, id)
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
	if err := h.TaskRepo.Update(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "data": t})
}

// Delete Task godoc
// @Summary Hapus task
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/tasks/{id} [delete]
func (h *Handlers) DeleteTask(c *gin.Context) {
	uid := c.GetString("user_id")
	id := c.Param("id")
	if err := h.TaskRepo.Delete(c.Request.Context(), uid, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response_code": http.StatusBadRequest, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response_code": http.StatusOK, "message": "deleted"})
}
