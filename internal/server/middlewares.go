package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/ashtishad/xm/common"
	"github.com/ashtishad/xm/internal/domain"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(gin.Logger())
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

// AuthMiddleware validates the JWT token with ecdsa public key from the Authorization header,
// extracts the user ID from claims, and fetches the corresponding user from the database.
// The authorizedUser is then stored in the request context for use in subsequent handlers.

// Parameters:
//   - userRepo: A UserRepository for fetching user data, passed as a parameter to allow for
//     dependency injection and easier testing.
func (s *Server) AuthMiddleware(userRepo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Missing authorization header"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid authorization header format"})
			c.Abort()
			return
		}

		accessToken := bearerToken[1]
		claims, err := s.Config.JWT.ValidateToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
			c.Abort()
			return
		}

		user, appErr := userRepo.FindBy(c.Request.Context(), common.DBColumnUUID, claims.UserID)
		if appErr != nil {
			c.JSON(appErr.Code(), ErrorResponse{Error: appErr.Error()})
			c.Abort()
			return
		}

		// Set the authorized user in the context
		c.Set("authorizedUser", user)
		c.Next()
	}
}
