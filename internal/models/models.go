package models

import (
	"github.com/google/uuid"
	"time"
)

const (
	ModeratorRole = "moderator"
	EmployeeRole  = "employee"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"password" db:"password_hash"`
	Role     string    `json:"role"`
}

type PVZ struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"created_at"`
	City             string    `json:"city"`
}

type Reception struct {
	ID       uuid.UUID `json:"id" db:"id"`
	DateTime time.Time `json:"dateTime" db:"created_at"`
	PVZID    uuid.UUID `json:"pvzId" db:"pickup_point_id"`
	Status   string    `json:"status"`
}

type Product struct {
	ID          uuid.UUID `json:"id" db:"id"`
	DateTime    time.Time `json:"dateTime" db:"created_at"`
	Type        string    `json:"type"`
	ReceptionId uuid.UUID `json:"receptionId"`
}
