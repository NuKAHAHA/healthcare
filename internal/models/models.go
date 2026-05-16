package models

import (
	"time"
)

const (
	RoleAdmin     = "admin"
	RoleRegistrar = "registrar"
	RoleDoctor    = "doctor"
)

// User represents a system user
type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
	// PasswordHash is never exposed in API responses
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Patient represents a patient record
type Patient struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	DateOfBirth  time.Time `json:"dateOfBirth"`
	Gender       string    `json:"gender"`
	Address      string    `json:"address"`
	MedicalInfo  string    `json:"medicalInfo"`
	RegisteredBy int64     `json:"registeredBy"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Appointment represents a doctor appointment
type Appointment struct {
	ID          int64     `json:"id"`
	PatientID   int64     `json:"patientId"`
	PatientName string    `json:"patientName"`
	DoctorID    int64     `json:"doctorId"`
	ScheduledAt time.Time `json:"scheduledAt"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"` // pending, completed, cancelled
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Treatment represents a treatment/prescription entry
type Treatment struct {
	ID            int64     `json:"id"`
	AppointmentID int64     `json:"appointmentId"`
	PatientID     int64     `json:"patientId"`
	DoctorID      int64     `json:"doctorId"`
	Diagnosis     string    `json:"diagnosis"`
	Prescription  string    `json:"prescription"`
	Notes         *string   `json:"notes"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Report represents a generated appointment report
type Report struct {
	ID            int64     `json:"id"`
	AppointmentID int64     `json:"appointmentId"`
	PatientID     int64     `json:"patientId"`
	DoctorID      int64     `json:"doctorId"`
	Diagnosis     string    `json:"diagnosis"`
	Prescription  string    `json:"prescription"`
	Notes         *string   `json:"notes"`
	GeneratedAt   time.Time `json:"generatedAt"`
}

// AuditLog represents a security audit log entry
type AuditLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"userId"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID int64     `json:"resourceId"`
	Status     string    `json:"status"` // success, failure
	Details    string    `json:"details"`
	CreatedAt  time.Time `json:"createdAt"`
}

type RefreshToken struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// Claims represents JWT token claims
type Claims struct {
	UserID int64  `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	// Standard claims are handled by jwt library
}

// ContextKey is used for storing values in request context
type ContextKey string

const (
	UserContextKey = ContextKey("user")
)

// AuthUser represents authenticated user context
type AuthUser struct {
	ID    int64
	Email string
	Role  string
}
