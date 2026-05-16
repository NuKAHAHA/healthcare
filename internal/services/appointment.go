package services

import (
	"context"
	"fmt"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"healthcare-api/internal/repositories"
	"time"
)

// AppointmentService handles appointment operations
type AppointmentService struct {
	appointmentRepo *repositories.AppointmentRepository
	patientRepo     *repositories.PatientRepository
	userRepo        *repositories.UserRepository
	auditRepo       *repositories.AuditLogRepository
	logger          *logger.Logger
}

// NewAppointmentService creates a new appointment service
func NewAppointmentService(
	appointmentRepo *repositories.AppointmentRepository,
	patientRepo *repositories.PatientRepository,
	userRepo *repositories.UserRepository,
	auditRepo *repositories.AuditLogRepository,
	logger *logger.Logger,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
		userRepo:        userRepo,
		auditRepo:       auditRepo,
		logger:          logger,
	}
}

// CreateAppointment creates a new appointment
// SECURITY: Registrars can create appointments, authorization checked in handler
func (s *AppointmentService) CreateAppointment(ctx context.Context, req *models.CreateAppointmentRequest, creatorID int64) (*models.Appointment, error) {
	// Input validation
	if req.PatientID == 0 || req.DoctorID == 0 {
		return nil, fmt.Errorf("patientId and doctorId are required")
	}

	if req.Reason == "" {
		return nil, fmt.Errorf("reason is required")
	}

	if len(req.Reason) > 500 {
		return nil, fmt.Errorf("reason too long")
	}

	// Validate patient exists
	patient, err := s.patientRepo.GetByID(ctx, req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("invalid patientId")
	}

	// Validate doctor exists
	doctor, err := s.userRepo.GetByID(ctx, req.DoctorID)
	if err != nil {
		return nil, fmt.Errorf("invalid doctorId")
	}

	// Validate doctor is actually a doctor
	if doctor.Role != models.RoleDoctor {
		return nil, fmt.Errorf("selected user is not a doctor")
	}

	// Parse scheduled time
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduledAt format (use RFC3339)")
	}

	// Validate scheduled time is in future
	if scheduledAt.Before(time.Now()) {
		return nil, fmt.Errorf("scheduledAt must be in the future")
	}

	// Create appointment
	appointment := &models.Appointment{
		PatientID:   req.PatientID,
		DoctorID:    req.DoctorID,
		ScheduledAt: scheduledAt,
		Reason:      req.Reason,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	appointment, err = s.appointmentRepo.Create(ctx, appointment)
	if err != nil {
		return nil, err
	}

	// Log appointment creation
	auditLog := &models.AuditLog{
		UserID:     creatorID,
		Action:     "appointment_creation",
		Resource:   "appointment",
		ResourceID: appointment.ID,
		Status:     "success",
		Details:    fmt.Sprintf("patient=%d, doctor=%d", patient.ID, doctor.ID),
		CreatedAt:  time.Now(),
	}
	_, _ = s.auditRepo.Create(ctx, auditLog)

	s.logger.InfoWithContext("appointment created", "appointment_creation", creatorID, fmt.Sprintf("patient=%d doctor=%d", patient.ID, doctor.ID))

	return appointment, nil
}

// GetAppointment retrieves an appointment by ID
func (s *AppointmentService) GetAppointment(ctx context.Context, appointmentID int64) (*models.Appointment, error) {
	return s.appointmentRepo.GetByID(ctx, appointmentID)
}

// UpdateStatus changes the status of an appointment (admin/registrar only via API).
func (s *AppointmentService) UpdateStatus(ctx context.Context, appointmentID int64, status string, requesterID int64) error {
	valid := map[string]bool{"pending": true, "completed": true, "cancelled": true}
	if !valid[status] {
		return fmt.Errorf("invalid status: must be pending, completed, or cancelled")
	}

	existing, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("appointment not found")
	}
	if existing.Status == "cancelled" {
		return fmt.Errorf("cannot change status of a cancelled appointment")
	}

	if err := s.appointmentRepo.UpdateStatus(ctx, appointmentID, status); err != nil {
		return err
	}

	auditLog := &models.AuditLog{
		UserID:     requesterID,
		Action:     "appointment_status_update",
		Resource:   "appointment",
		ResourceID: appointmentID,
		Status:     "success",
		Details:    fmt.Sprintf("status=%s", status),
		CreatedAt:  time.Now(),
	}
	_, _ = s.auditRepo.Create(ctx, auditLog)

	s.logger.InfoWithContext("appointment status updated", "appointment_status_update", requesterID,
		fmt.Sprintf("appointment=%d status=%s", appointmentID, status))
	return nil
}

// GetAppointmentsByPatient retrieves paginated appointments for a patient
func (s *AppointmentService) GetAppointmentsByPatient(ctx context.Context, patientID int64, page, pageSize int) ([]*models.Appointment, error) {
	return s.appointmentRepo.GetByPatientID(ctx, patientID, page, pageSize)
}

// GetAppointmentsByDoctor retrieves paginated appointments for a doctor
func (s *AppointmentService) GetAppointmentsByDoctor(ctx context.Context, doctorID int64, page, pageSize int) ([]*models.Appointment, error) {
	return s.appointmentRepo.GetByDoctorID(ctx, doctorID, page, pageSize)
}
