package repositories

import (
	"context"
	"fmt"
	"healthcare-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppointmentRepository struct {
	db *pgxpool.Pool
}

func NewAppointmentRepository(db *pgxpool.Pool) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(ctx context.Context, appointment *models.Appointment) (*models.Appointment, error) {
	query := `
		INSERT INTO appointments (patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', NOW(), NOW())
		RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, notes, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		appointment.PatientID,
		appointment.DoctorID,
		appointment.ScheduledAt,
		appointment.Reason,
	).Scan(
		&appointment.ID, &appointment.PatientID, &appointment.DoctorID,
		&appointment.ScheduledAt, &appointment.Reason, &appointment.Status,
		&appointment.Notes, &appointment.CreatedAt, &appointment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}
	return appointment, nil
}

func (r *AppointmentRepository) GetByID(ctx context.Context, id int64) (*models.Appointment, error) {
	appointment := &models.Appointment{}
	query := `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, notes, created_at, updated_at
		FROM appointments WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&appointment.ID, &appointment.PatientID, &appointment.DoctorID,
		&appointment.ScheduledAt, &appointment.Reason, &appointment.Status,
		&appointment.Notes, &appointment.CreatedAt, &appointment.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("appointment not found")
		}
		return nil, fmt.Errorf("failed to fetch appointment: %w", err)
	}
	return appointment, nil
}

// GetAll retrieves all appointments with patient name (admin/registrar view).
func (r *AppointmentRepository) GetAll(ctx context.Context, page, pageSize int) ([]*models.Appointment, error) {
	page, pageSize = clampPagination(page, pageSize)
	offset := (page - 1) * pageSize
	query := `
		SELECT a.id, a.patient_id,
		       p.first_name || ' ' || p.last_name AS patient_name,
		       a.doctor_id, a.scheduled_at, a.reason, a.status, a.notes, a.created_at, a.updated_at
		FROM appointments a
		LEFT JOIN patients p ON p.id = a.patient_id
		ORDER BY a.scheduled_at DESC
		LIMIT $1 OFFSET $2
	`
	return r.scanAppointments(ctx, query, pageSize, offset)
}

func (r *AppointmentRepository) GetByPatientID(ctx context.Context, patientID int64, page, pageSize int) ([]*models.Appointment, error) {
	page, pageSize = clampPagination(page, pageSize)
	offset := (page - 1) * pageSize
	query := `
		SELECT a.id, a.patient_id,
		       p.first_name || ' ' || p.last_name AS patient_name,
		       a.doctor_id, a.scheduled_at, a.reason, a.status, a.notes, a.created_at, a.updated_at
		FROM appointments a
		LEFT JOIN patients p ON p.id = a.patient_id
		WHERE a.patient_id = $1 ORDER BY a.scheduled_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.scanAppointments(ctx, query, patientID, pageSize, offset)
}

func (r *AppointmentRepository) GetByDoctorID(ctx context.Context, doctorID int64, page, pageSize int) ([]*models.Appointment, error) {
	page, pageSize = clampPagination(page, pageSize)
	offset := (page - 1) * pageSize
	query := `
		SELECT a.id, a.patient_id,
		       p.first_name || ' ' || p.last_name AS patient_name,
		       a.doctor_id, a.scheduled_at, a.reason, a.status, a.notes, a.created_at, a.updated_at
		FROM appointments a
		LEFT JOIN patients p ON p.id = a.patient_id
		WHERE a.doctor_id = $1 ORDER BY a.scheduled_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.scanAppointments(ctx, query, doctorID, pageSize, offset)
}

func (r *AppointmentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	result, err := r.db.Exec(ctx,
		`UPDATE appointments SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update appointment status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("appointment not found")
	}
	return nil
}

func (r *AppointmentRepository) HasAppointmentWithPatient(ctx context.Context, doctorID, patientID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM appointments WHERE doctor_id = $1 AND patient_id = $2`,
		doctorID, patientID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check appointment relationship: %w", err)
	}
	return count > 0, nil
}

func (r *AppointmentRepository) scanAppointments(ctx context.Context, query string, args ...interface{}) ([]*models.Appointment, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch appointments: %w", err)
	}
	defer rows.Close()

	var appointments []*models.Appointment
	for rows.Next() {
		a := &models.Appointment{}
		if err := rows.Scan(
			&a.ID, &a.PatientID, &a.PatientName,
			&a.DoctorID, &a.ScheduledAt, &a.Reason, &a.Status,
			&a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan appointment: %w", err)
		}
		appointments = append(appointments, a)
	}
	return appointments, rows.Err()
}
