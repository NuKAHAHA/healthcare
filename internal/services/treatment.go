package services

import (
	"context"
	"fmt"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"healthcare-api/internal/repositories"
	"time"
)

// TreatmentService handles treatment operations
type TreatmentService struct {
	treatmentRepo   *repositories.TreatmentRepository
	appointmentRepo *repositories.AppointmentRepository
	patientRepo     *repositories.PatientRepository
	auditRepo       *repositories.AuditLogRepository
	logger          *logger.Logger
}

// NewTreatmentService creates a new treatment service
func NewTreatmentService(
	treatmentRepo *repositories.TreatmentRepository,
	appointmentRepo *repositories.AppointmentRepository,
	patientRepo *repositories.PatientRepository,
	auditRepo *repositories.AuditLogRepository,
	logger *logger.Logger,
) *TreatmentService {
	return &TreatmentService{
		treatmentRepo:   treatmentRepo,
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
		auditRepo:       auditRepo,
		logger:          logger,
	}
}

// AddTreatment adds a treatment/prescription to an appointment
// SECURITY: Only the assigned doctor can add treatment, authorization checked in handler
func (s *TreatmentService) AddTreatment(ctx context.Context, req *models.CreateTreatmentRequest, doctorID int64) (*models.Treatment, error) {
	// Input validation
	if req.AppointmentID == 0 {
		return nil, fmt.Errorf("appointmentId is required")
	}

	if req.Diagnosis == "" {
		return nil, fmt.Errorf("diagnosis is required")
	}

	if len(req.Diagnosis) > 1000 {
		return nil, fmt.Errorf("diagnosis too long")
	}

	if len(req.Prescription) > 1000 {
		return nil, fmt.Errorf("prescription too long")
	}

	// Validate appointment exists
	appointment, err := s.appointmentRepo.GetByID(ctx, req.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid appointmentId")
	}

	// SECURITY: Verify doctor owns this appointment
	if appointment.DoctorID != doctorID {
		s.logger.WarnWithContext("unauthorized treatment attempt", "treatment_add", doctorID, fmt.Sprintf("appointment=%d", req.AppointmentID))
		return nil, fmt.Errorf("not authorized to add treatment to this appointment")
	}

	if appointment.Status == "cancelled" {
		return nil, fmt.Errorf("cannot add treatment to a cancelled appointment")
	}

	// Validate patient exists
	patient, err := s.patientRepo.GetByID(ctx, appointment.PatientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found")
	}

	// Create treatment
	treatment := &models.Treatment{
		AppointmentID: req.AppointmentID,
		PatientID:     appointment.PatientID,
		DoctorID:      doctorID,
		Diagnosis:     req.Diagnosis,
		Prescription:  req.Prescription,
		Notes:         req.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	treatment, err = s.treatmentRepo.Create(ctx, treatment)
	if err != nil {
		return nil, err
	}

	// Mark the appointment completed now that a treatment has been recorded.
	_ = s.appointmentRepo.UpdateStatus(ctx, req.AppointmentID, "completed")

	// Log treatment entry (do not log actual diagnosis/prescription for privacy)
	auditLog := &models.AuditLog{
		UserID:     doctorID,
		Action:     "treatment_add",
		Resource:   "treatment",
		ResourceID: treatment.ID,
		Status:     "success",
		Details:    fmt.Sprintf("appointment=%d, patient=%d", appointment.ID, patient.ID),
		CreatedAt:  time.Now(),
	}
	_, _ = s.auditRepo.Create(ctx, auditLog)

	s.logger.InfoWithContext("treatment added", "treatment_add", doctorID, fmt.Sprintf("appointment=%d patient=%d", appointment.ID, patient.ID))

	return treatment, nil
}

// GetTreatment retrieves treatment for an appointment
func (s *TreatmentService) GetTreatment(ctx context.Context, appointmentID int64) (*models.Treatment, error) {
	return s.treatmentRepo.GetByAppointmentID(ctx, appointmentID)
}

// GenerateReport generates a report from a treatment
func (s *TreatmentService) GenerateReport(ctx context.Context, appointmentID int64, requesterID int64, requesterRole string) (*models.Report, error) {
	// Validate appointment exists
	appointment, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("appointment not found")
	}

	if requesterRole != models.RoleAdmin && appointment.DoctorID != requesterID {
		s.logger.WarnWithContext("unauthorized report access", "report_generation", requesterID, fmt.Sprintf("appointment=%d", appointmentID))
		return nil, fmt.Errorf("not authorized to access this report")
	}

	// Get treatment for this appointment
	treatment, err := s.treatmentRepo.GetByAppointmentID(ctx, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("no treatment found for this appointment")
	}

	// Create report (ID = treatment ID since report is not stored separately)
	report := &models.Report{
		ID:            treatment.ID,
		AppointmentID: appointmentID,
		PatientID:     appointment.PatientID,
		DoctorID:      appointment.DoctorID,
		Diagnosis:     treatment.Diagnosis,
		Prescription:  treatment.Prescription,
		Notes:         treatment.Notes,
		GeneratedAt:   time.Now(),
	}

	// Log report generation (do not log medical data)
	auditLog := &models.AuditLog{
		UserID:     requesterID,
		Action:     "report_generation",
		Resource:   "report",
		ResourceID: appointmentID,
		Status:     "success",
		Details:    fmt.Sprintf("appointment=%d, patient=%d", appointmentID, appointment.PatientID),
		CreatedAt:  time.Now(),
	}
	_, _ = s.auditRepo.Create(ctx, auditLog)

	s.logger.InfoWithContext("report generated", "report_generation", requesterID, fmt.Sprintf("appointment=%d", appointmentID))

	return report, nil
}
