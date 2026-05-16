package repositories

import (
	"context"
	"fmt"
	"healthcare-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TreatmentRepository handles treatment database operations
type TreatmentRepository struct {
	db *pgxpool.Pool
}

// NewTreatmentRepository creates a new treatment repository
func NewTreatmentRepository(db *pgxpool.Pool) *TreatmentRepository {
	return &TreatmentRepository{db: db}
}

// Create creates a new treatment using parameterized query
func (r *TreatmentRepository) Create(ctx context.Context, treatment *models.Treatment) (*models.Treatment, error) {
	// SECURITY: Using parameterized query to prevent SQL injection
	query := `
		INSERT INTO treatments (appointment_id, patient_id, doctor_id, diagnosis, prescription, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, appointment_id, patient_id, doctor_id, diagnosis, prescription, notes, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		treatment.AppointmentID,
		treatment.PatientID,
		treatment.DoctorID,
		treatment.Diagnosis,
		treatment.Prescription,
		treatment.Notes,
	).Scan(
		&treatment.ID,
		&treatment.AppointmentID,
		&treatment.PatientID,
		&treatment.DoctorID,
		&treatment.Diagnosis,
		&treatment.Prescription,
		&treatment.Notes,
		&treatment.CreatedAt,
		&treatment.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create treatment: %w", err)
	}

	return treatment, nil
}

// GetByAppointmentID retrieves treatment for an appointment using parameterized query
func (r *TreatmentRepository) GetByAppointmentID(ctx context.Context, appointmentID int64) (*models.Treatment, error) {
	treatment := &models.Treatment{}

	// SECURITY: Using parameterized query to prevent SQL injection
	query := `
		SELECT id, appointment_id, patient_id, doctor_id, diagnosis, prescription, notes, created_at, updated_at
		FROM treatments WHERE appointment_id = $1
		ORDER BY created_at DESC LIMIT 1
	`

	err := r.db.QueryRow(ctx, query, appointmentID).Scan(
		&treatment.ID,
		&treatment.AppointmentID,
		&treatment.PatientID,
		&treatment.DoctorID,
		&treatment.Diagnosis,
		&treatment.Prescription,
		&treatment.Notes,
		&treatment.CreatedAt,
		&treatment.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("treatment not found")
		}
		return nil, fmt.Errorf("failed to fetch treatment: %w", err)
	}

	return treatment, nil
}
