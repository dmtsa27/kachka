package api

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/dmtsa27/kachka.git/docs"
	"github.com/dmtsa27/kachka.git/pkg/domain"
	"github.com/dmtsa27/kachka.git/pkg/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewServer,
	),
	fx.Invoke(
		func(lc fx.Lifecycle, s *Server) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return s.Run(ctx)
				},
			})
		},
	),
)

type Server struct {
	router  *gin.Engine
	service *service.Service
}

func NewServer(svc *service.Service) *Server {
	router := gin.Default()
	s := &Server{
		router:  router,
		service: svc,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.router.GET("/ping", s.handlePing)
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	admin := s.router.Group("/admin")
	{
		admin.GET("/stats", s.handleGetStats)

		// User CRUD
		admin.GET("/users", s.handleGetUsers)
		admin.POST("/users", s.handleCreateUser)
		admin.GET("/users/:id", s.handleGetUser)
		admin.PUT("/users/:id", s.handleUpdateUser)
		admin.DELETE("/users/:id", s.handleDeleteUser)

		// Challenge CRUD
		admin.GET("/challenges", s.handleGetChallenges)
		admin.GET("/challenges/:id", s.handleGetChallenge)
		admin.DELETE("/challenges/:id", s.handleDeleteChallenge)

		admin.POST("/workouts", s.handleAddWorkout)
	}
}

// handleGetUser godoc
// @Summary Get user by ID
// @Tags admin
// @Param id path int64 true "Telegram User ID"
// @Produce json
// @Success 200 {object} domain.User
// @Router /admin/users/{id} [get]
func (s *Server) handleGetUser(c *gin.Context) {
	idStr := c.Param("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)
	user, err := s.service.ReadUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// handleCreateUser godoc
// @Summary Create or Upsert user
// @Tags admin
// @Accept json
// @Produce json
// @Param user body domain.User true "User model"
// @Success 201 {object} domain.User
// @Router /admin/users [post]
func (s *Server) handleCreateUser(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.service.RegisterUser(c.Request.Context(), user.TelegramID, user.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// handleUpdateUser godoc
// @Summary Update existing user
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int64 true "Telegram User ID"
// @Param user body domain.User true "User model"
// @Success 200 {object} domain.User
// @Router /admin/users/{id} [put]
func (s *Server) handleUpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.TelegramID = id
	if err := s.service.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// handleDeleteUser godoc
// @Summary Delete user and related data
// @Tags admin
// @Param id path int64 true "Telegram User ID"
// @Success 204
// @Router /admin/users/{id} [delete]
func (s *Server) handleDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)
	if err := s.service.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// handleGetChallenges godoc
// @Summary List all active challenges
// @Tags admin
// @Produce json
// @Success 200 {array} domain.Challenge
// @Router /admin/challenges [get]
func (s *Server) handleGetChallenges(c *gin.Context) {
	challenges, err := s.service.ActiveChallenges(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, challenges)
}

// handleGetChallenge godoc
// @Summary Get challenge by ID
// @Tags admin
// @Param id path int true "Challenge ID"
// @Produce json
// @Success 200 {object} domain.Challenge
// @Router /admin/challenges/{id} [get]
func (s *Server) handleGetChallenge(c *gin.Context) {
	idStr := c.Param("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	ch, err := s.service.GetChallenge(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}
	c.JSON(http.StatusOK, ch)
}

// handleDeleteChallenge godoc
// @Summary Delete challenge
// @Tags admin
// @Param id path int true "Challenge ID"
// @Success 204
// @Router /admin/challenges/{id} [delete]
func (s *Server) handleDeleteChallenge(c *gin.Context) {
	idStr := c.Param("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	if err := s.service.DeleteChallenge(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// handleGetUsers godoc
// @Summary Get all active users
// @Tags admin
// @Produce json
// @Success 200 {array} domain.User
// @Router /admin/users [get]
func (s *Server) handleGetUsers(c *gin.Context) {
	users, err := s.service.GetAllActiveUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// handlePing godoc
// @Summary Ping-pong
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func (s *Server) handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// handleGetStats godoc
// @Summary Get statistics for a chat
// @Tags admin
// @Param chat_id query int64 true "Chat ID"
// @Produce json
// @Success 200 {array} domain.UserStats
// @Router /admin/stats [get]
func (s *Server) handleGetStats(c *gin.Context) {
	chatIDStr := c.Query("chat_id")
	var chatID int64
	fmt.Sscanf(chatIDStr, "%d", &chatID)

	stats, err := s.service.GetStats(c.Request.Context(), chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

type AddWorkoutRequest struct {
	Username string `json:"username" binding:"required"`
	ChatID   int64  `json:"chat_id" binding:"required"`
	Amount   int    `json:"amount" binding:"required"`
}

// handleAddWorkout godoc
// @Summary Add manual workouts for a user
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AddWorkoutRequest true "Request body"
// @Success 200 {object} map[string]interface{}
// @Router /admin/workouts [post]
func (s *Server) handleAddWorkout(c *gin.Context) {
	var req AddWorkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	added, err := s.service.AddWorkoutDirect(c.Request.Context(), req.ChatID, req.Username, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added workouts via admin", "count": added})
}

func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: s.router,
	}

	go func() {
		fmt.Println("Starting API server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("API server error: %v\n", err)
		}
	}()

	return nil
}
