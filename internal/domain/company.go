package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	CompanyTypeCorporation        = "Corporations"
	CompanyTypeNonProfit          = "NonProfit"
	CompanyTypeCooperative        = "Cooperative"
	CompanyTypeSoleProprietorship = "Sole Proprietorship"
)

type Company struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Description       *string    `json:"description,omitempty"`
	AmountOfEmployees int        `json:"amountOfEmployees"`
	Registered        bool       `json:"registered"`
	Type              string     `json:"type"`
	CreatedAt         *time.Time `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt"`
}
