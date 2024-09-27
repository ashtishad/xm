package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ashtishad/xm/common"
	"github.com/ashtishad/xm/internal/domain"
	"github.com/ashtishad/xm/internal/security"
)

type UserHandler struct {
	userRepo   domain.UserRepository
	JWTManager *security.JWTManager
	l          *slog.Logger
}

func NewAuthHandler(userRepo domain.UserRepository, jwtManager *security.JWTManager, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		JWTManager: jwtManager,
		l:          logger,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Hashes password using bcrypt before storage.
// @Description Generates JWT access token using ECDSA encryption.
// @Description Sets HTTP-only cookie with access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RegisterUserRequest true "User registration details"
// @Success 201 {object} RegisterUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error(common.ErrInvalidRequest, "err", formatValidationError(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: formatValidationError(err)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.TimeOutRegisterUser)
	defer cancel()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.l.Error("unable to hash password", "err", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{common.ErrUnexpectedServer})
		return
	}

	now := time.Now().UTC()
	user := &domain.User{
		UUID:         uuid.New(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPass),
		Status:       domain.UserStatusActive,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	createdUser, appErr := h.userRepo.Create(ctx, user)
	if appErr != nil {
		c.JSON(appErr.Code(), ErrorResponse{appErr.Error()})
		return
	}

	accessToken, err := h.JWTManager.GenerateAccessToken(
		createdUser.UUID.String(),
		h.JWTManager.AccessExp,
	)

	if err != nil {
		h.l.Error("failed to generate access token", "err", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{common.ErrUnexpectedServer})
		return
	}

	c.SetCookie("accessToken", accessToken, int(h.JWTManager.AccessExp.Seconds()), "/", "", true, true)

	resp := RegisterUserResponse{
		User: *createdUser,
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Authenticate a user and provide access token
// @Description Verifies password using bcrypt comparison.
// @Description Generates new JWT access token using ECDSA encryption.
// @Description Sets HTTP-only cookie with new access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "User login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error(common.ErrInvalidRequest, "err", formatValidationError(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: formatValidationError(err)})
		return
	}

	user, appErr := h.userRepo.FindBy(c.Request.Context(), common.DBColumnEmail, req.Email)
	if appErr != nil {
		c.JSON(appErr.Code(), gin.H{"error": appErr.Error()})
		return
	}

	if !verifyPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{common.ErrIncorrectPassword})
		return
	}

	accessToken, err := h.JWTManager.GenerateAccessToken(
		user.UUID.String(),
		h.JWTManager.AccessExp,
	)

	if err != nil {
		h.l.Error("failed to generate access token", "err", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{common.ErrUnexpectedServer})
		return
	}

	c.SetCookie("accessToken", accessToken, int(h.JWTManager.AccessExp.Seconds()), "/", "", true, true)

	resp := LoginResponse{
		User: *user,
	}

	c.JSON(http.StatusOK, resp)
}

func verifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
