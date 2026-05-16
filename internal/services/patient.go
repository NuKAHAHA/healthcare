package services

import (
	"context"
	"fmt"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"healthcare-api/internal/repositories"
	"regexp"
	"time"
)

// E.164 phone: +<1-9><7-14 digits>  (total 8-15 chars)
var (
	emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegexp = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
)

type PatientService struct {
	patientRepo *repositories.PatientRepository
	auditRepo   *repositories.AuditLogRepository
	logger      *logger.Logger
}

func NewPatientService(
	patientRepo *repositories.PatientRepository,
	auditRepo *repositories.AuditLogRepository,
	appLogger *logger.Logger,
) *PatientService {
	return &PatientService{
		patientRepo: patientRepo,
		auditRepo:   auditRepo,
		logger:      appLogger,
	}
}

// RegisterPatient validates input and creates a new patient record.
func (s *PatientService) RegisterPatient(ctx context.Context, req *models.CreatePatientRequest, registrarID int64) (*models.Patient, error) {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return nil, fmt.Errorf("firstName, lastName, and email are required")
	}
	if len(req.FirstName) > 100 {
		return nil, fmt.Errorf("firstName must be at most 100 characters")
	}
	if len(req.LastName) > 100 {
		return nil, fmt.Errorf("lastName must be at most 100 characters")
	}
	if len(req.Email) > 255 {
		return nil, fmt.Errorf("email must be at most 255 characters")
	}

	// Email format validation
	if !emailRegexp.MatchString(req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Phone format validation (optional field, but must be E.164 if provided)
	if req.Phone != "" && !phoneRegexp.MatchString(req.Phone) {
		return nil, fmt.Errorf("invalid phone format — use E.164 (e.g. +12025551234)")
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, fmt.Errorf("invalid dateOfBirth format — use YYYY-MM-DD")
	}
	if time.Since(dob) < 0 {
		return nil, fmt.Errorf("dateOfBirth cannot be in the future")
	}

	if req.Gender != "" && req.Gender != "M" && req.Gender != "F" && req.Gender != "O" {
		return nil, fmt.Errorf("gender must be M, F, or O")
	}

	patient, err := s.patientRepo.Create(ctx, &models.Patient{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		DateOfBirth:  dob,
		Gender:       req.Gender,
		Address:      req.Address,
		MedicalInfo:  req.MedicalInfo,
		RegisteredBy: registrarID,
	})
	if err != nil {
		return nil, err
	}

	s.writeAudit(ctx, registrarID, "patient_registration", "patient", patient.ID, "success",
		fmt.Sprintf("patient_id=%d", patient.ID))
	s.logger.InfoWithContext("patient registered", "patient_registration", registrarID,
		fmt.Sprintf("patient_id=%d", patient.ID))

	return patient, nil
}

// GetPatient retrieves a single patient. Authorization is enforced in the handler.
func (s *PatientService) GetPatient(ctx context.Context, patientID int64) (*models.Patient, error) {
	return s.patientRepo.GetByID(ctx, patientID)
}

// GetAllPatients retrieves a paginated list of patients.
func (s *PatientService) GetAllPatients(ctx context.Context, page, pageSize int) ([]*models.Patient, error) {
	return s.patientRepo.GetAll(ctx, page, pageSize)
}

// GetPatientsByDoctor retrieves patients who have appointments with the given doctor.
func (s *PatientService) GetPatientsByDoctor(ctx context.Context, doctorID int64, page, pageSize int) ([]*models.Patient, error) {
	return s.patientRepo.GetByDoctorID(ctx, doctorID, page, pageSize)
}

// LogPatientAccess writes an audit log for a patient read event.
func (s *PatientService) LogPatientAccess(ctx context.Context, accessorID, patientID int64, ip string) {
	s.writeAudit(ctx, accessorID, "patient_access", "patient", patientID, "success",
		fmt.Sprintf("accessor_id=%d patient_id=%d ip=%s", accessorID, patientID, ip))
}

func (s *PatientService) writeAudit(ctx context.Context, userID int64, action, resource string, resourceID int64, status, details string) {
	_, _ = s.auditRepo.Create(ctx, &models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
		Details:    details,
		CreatedAt:  time.Now(),
	})
}
