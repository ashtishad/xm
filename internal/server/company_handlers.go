package server

import (
	"log/slog"
	"net/http"

	"github.com/ashtishad/xm/common"
	"github.com/ashtishad/xm/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	companyRepo domain.CompanyRepository
	l           *slog.Logger
}

func NewCompanyHandler(companyRepo domain.CompanyRepository, logger *slog.Logger) *CompanyHandler {
	return &CompanyHandler{
		companyRepo: companyRepo,
		l:           logger,
	}
}

// CreateCompany godoc
// @Summary Create a new company
// @Description Creates a new company with the provided details
// @Tags companies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param input body CreateCompanyRequest true "Company creation details"
// @Success 201 {object} domain.Company
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /companies [post]
func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error(common.ErrInvalidRequest, "err", formatValidationError(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: formatValidationError(err)})
		return
	}

	company := &domain.Company{
		ID:                uuid.New(),
		Name:              req.Name,
		Description:       req.Description,
		AmountOfEmployees: req.AmountOfEmployees,
		Registered:        req.Registered,
		Type:              req.Type,
	}

	createdCompany, appErr := h.companyRepo.Create(c.Request.Context(), company)
	if appErr != nil {
		c.JSON(appErr.Code(), ErrorResponse{Error: appErr.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdCompany)
}

// GetCompany godoc
// @Summary Get a company by ID(UUID)
// @Description Retrieves a company's details by its ID
// @Tags companies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Company ID"
// @Success 200 {object} domain.Company
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /companies/{id} [get]
func (h *CompanyHandler) GetCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid company ID"})
		return
	}

	company, appErr := h.companyRepo.FindByID(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(appErr.Code(), ErrorResponse{Error: appErr.Error()})
		return
	}

	c.JSON(http.StatusOK, company)
}

// UpdateCompany godoc
// @Summary Update a company
// @Description Updates a company's details
// @Tags companies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Company ID"
// @Param input body UpdateCompanyRequest true "Company update details"
// @Success 200 {object} domain.Company
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /companies/{id} [patch]
func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid company ID"})
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error(common.ErrInvalidRequest, "err", formatValidationError(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: formatValidationError(err)})
		return
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.AmountOfEmployees != nil {
		updates["amount_of_employees"] = *req.AmountOfEmployees
	}

	if req.Registered != nil {
		updates["registered"] = *req.Registered
	}

	if req.Type != nil {
		updates["type"] = *req.Type
	}

	updatedCompany, appErr := h.companyRepo.Update(c.Request.Context(), id, updates)
	if appErr != nil {
		c.JSON(appErr.Code(), ErrorResponse{Error: appErr.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedCompany)
}

// DeleteCompany godoc
// @Summary Delete a company
// @Description Soft deletes a company by setting its deleted_at timestamp
// @Tags companies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Company ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /companies/{id} [delete]
func (h *CompanyHandler) DeleteCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid company ID"})
		return
	}

	appErr := h.companyRepo.Delete(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(appErr.Code(), ErrorResponse{Error: appErr.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
