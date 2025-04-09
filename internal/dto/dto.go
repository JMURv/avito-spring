package dto

import (
	md "github.com/JMURv/avito-spring/internal/models"
	"github.com/google/uuid"
	"time"
)

type DummyLoginRequest struct {
	Role string `json:"role" validate:"required"`
}

type DummyLoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type RegisterResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type CreatePVZRequest struct {
	City string `json:"city" validate:"required"`
}

type CreatePVZResponse struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"created_at"`
	City             string    `json:"city"`
}

type GetPVZResponse struct {
	PVZ        md.PVZ `json:"pvz"`
	Receptions []struct {
		Reception md.Reception `json:"reception"`
		Products  []md.Product `json:"products"`
	} `json:"receptions"`
}

type CreateReceptionRequest struct {
	PVZID uuid.UUID `json:"pvzId" validate:"required"`
}

type CreateReceptionResponse struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime" db:"created_at"`
	PVZID    uuid.UUID `json:"pvzId" db:"pickup_point_id"`
	Status   string    `json:"status"`
}

type AddItemRequest struct {
	Type  string    `json:"type" validate:"required"`
	PVZID uuid.UUID `json:"pvzId" validate:"required"`
}

type AddItemResponse struct {
	ID          uuid.UUID `json:"id"`
	DateTime    time.Time `json:"dateTime" db:"created_at"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionId" db:"reception_id"`
}
