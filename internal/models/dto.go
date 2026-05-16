package models

import "time"

// DTOs separate API contracts from database models to prevent data leakage

// LoginRequest is the request for authentication
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the request for user registration
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
}

// LoginResponse is returned after successful login
type LoginResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	User         UserDTO   `json:"user"`
}

// UserDTO is the response DTO for user - no password hash exposed
type UserDTO struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
}

// CreatePatientRequest is the request to register a new patient
type CreatePatientRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"dateOfBirth"` // YYYY-MM-DD
	Gender      string `json:"gender"`      // M, F
	Address     string `json:"address"`
	MedicalInfo string `json:"medicalInfo"`
}

// PatientDTO is the response DTO for patient
type PatientDTO struct {
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

// CreateAppointmentRequest is the request to create an appointment
type CreateAppointmentRequest struct {
	PatientID   int64  `json:"patientId"`
	DoctorID    int64  `json:"doctorId"`
	ScheduledAt string `json:"scheduledAt"` // RFC3339 format
	Reason      string `json:"reason"`
}

// AppointmentDTO is the response DTO for appointment
type AppointmentDTO struct {
	ID          int64     `json:"id"`
	PatientID   int64     `json:"patientId"`
	PatientName string    `json:"patientName"`
	DoctorID    int64     `json:"doctorId"`
	ScheduledAt time.Time `json:"scheduledAt"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// UpdateAppointmentStatusRequest is the request to change appointment status
type UpdateAppointmentStatusRequest struct {
	Status string `json:"status"` // pending, completed, cancelled
}

// CreateTreatmentRequest is the request to add a treatment
type CreateTreatmentRequest struct {
	AppointmentID int64   `json:"appointmentId"`
	Diagnosis     string  `json:"diagnosis"`
	Prescription  string  `json:"prescription"`
	Notes         *string `json:"notes"`
}

// TreatmentDTO is the response DTO for treatment
type TreatmentDTO struct {
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

// ReportDTO is the response DTO for report
type ReportDTO struct {
	ID            int64     `json:"id"`
	AppointmentID int64     `json:"appointmentId"`
	PatientID     int64     `json:"patientId"`
	DoctorID      int64     `json:"doctorId"`
	Diagnosis     string    `json:"diagnosis"`
	Prescription  string    `json:"prescription"`
	Notes         *string   `json:"notes"`
	GeneratedAt   time.Time `json:"generatedAt"`
}

// AuditLogDTO is the response DTO for audit logs (admin only)
type AuditLogDTO struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"userId"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID int64     `json:"resourceId"`
	Status     string    `json:"status"`
	Details    string    `json:"details"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
