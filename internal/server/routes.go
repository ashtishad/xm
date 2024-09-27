package server

import (
	"github.com/ashtishad/xm/internal/domain"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) setupRoutes() {
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.router.Group("/api")
	api.GET("/health", s.dbHealthHandler)

	userRepo := domain.NewUserRepository(s.db, s.Logger)
	companyRepo := domain.NewCompanyRepository(s.db, s.Logger)

	s.registerAuthRoutes(api, userRepo)
	s.registerCompanyRoutes(api, companyRepo, userRepo)
}

func (s *Server) registerAuthRoutes(rg *gin.RouterGroup, userRepo domain.UserRepository) {
	authHandler := NewAuthHandler(userRepo, s.Config.JWT, s.Logger)

	rg.POST("/register", authHandler.Register)
	rg.POST("/login", authHandler.Login)
}

func (s *Server) registerCompanyRoutes(rg *gin.RouterGroup, companyRepo domain.CompanyRepository, userRepo domain.UserRepository) {
	companyHandler := NewCompanyHandler(companyRepo, s.Logger)

	companies := rg.Group("/companies")

	companies.Use(s.AuthMiddleware(userRepo))
	{
		companies.POST("/", companyHandler.CreateCompany)
		companies.GET("/:id", companyHandler.GetCompany)
		companies.PATCH("/:id", companyHandler.UpdateCompany)
		companies.DELETE("/:id", companyHandler.DeleteCompany)
	}
}
