package repositories

import (
	"context"
	"fmt"
	"healthcare-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PatientRepository handles patient database operations
type PatientRepository struct {
	db *pgxpool.Pool
}

// NewPatientRepository creates a new patient repository
func NewPatientRepository(db *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{db: db}
}

// Create creates a new patient using parameterized query
func (r *PatientRepository) Create(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	// SECURITY: Using parameterized query to prevent SQL injection
	query := `
		INSERT INTO patients (first_name, last_name, email, phone, date_of_birth, gender, address, medical_info, registered_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, first_name, last_name, email, phone, date_of_birth, gender, address, medical_info, registered_by, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		patient.FirstName,
		patient.LastName,
		patient.Email,
		patient.Phone,
		patient.DateOfBirth,
		patient.Gender,
		patient.Address,
		patient.MedicalInfo,
		patient.RegisteredBy,
	).Scan(
		&patient.ID,
		&patient.FirstName,
		&patient.LastName,
		&patient.Email,
		&patient.Phone,
		&patient.DateOfBirth,
		&patient.Gender,
		&patient.Address,
		&patient.MedicalInfo,
		&patient.RegisteredBy,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create patient: %w", err)
	}

	return patient, nil
}

// GetByID retrieves a patient by ID using parameterized query
func (r *PatientRepository) GetByID(ctx context.Context, id int64) (*models.Patient, error) {
	patient := &models.Patient{}

	// SECURITY: Using parameterized query to prevent SQL injection
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, address, medical_info, registered_by, created_at, updated_at
		FROM patients WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&patient.ID,
		&patient.FirstName,
		&patient.LastName,
		&patient.Email,
		&patient.Phone,
		&patient.DateOfBirth,
		&patient.Gender,
		&patient.Address,
		&patient.MedicalInfo,
		&patient.RegisteredBy,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, fmt.Errorf("failed to fetch patient: %w", err)
	}

	return patient, nil
}

// GetByDoctorID returns distinct patients who have at least one appointment with the given doctor.
func (r *PatientRepository) GetByDoctorID(ctx context.Context, doctorID int64, page, pageSize int) ([]*models.Patient, error) {
	page, pageSize = clampPagination(page, pageSize)
	offset := (page - 1) * pageSize

	query := `
		SELECT DISTINCT ON (p.id)
		       p.id, p.first_name, p.last_name, p.email, p.phone,
		       p.date_of_birth, p.gender, p.address, p.medical_info,
		       p.registered_by, p.created_at, p.updated_at
		FROM patients p
		INNER JOIN appointments a ON a.patient_id = p.id
		WHERE a.doctor_id = $1
		ORDER BY p.id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, doctorID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch doctor patients: %w", err)
	}
	defer rows.Close()

	var patients []*models.Patient
	for rows.Next() {
		p := &models.Patient{}
		if err := rows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
			&p.DateOfBirth, &p.Gender, &p.Address, &p.MedicalInfo,
			&p.RegisteredBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan patient: %w", err)
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func clampPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// GetAll retrieves all patients with pagination (default page=1, pageSize=50, max pageSize=100)
func (r *PatientRepository) GetAll(ctx context.Context, page, pageSize int) ([]*models.Patient, error) {
	page, pageSize = clampPagination(page, pageSize)
	offset := (page - 1) * pageSize

	var patients []*models.Patient

	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, address, medical_info, registered_by, created_at, updated_at
		FROM patients ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch patients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		patient := &models.Patient{}
		err := rows.Scan(
			&patient.ID,
			&patient.FirstName,
			&patient.LastName,
			&patient.Email,
			&patient.Phone,
			&patient.DateOfBirth,
			&patient.Gender,
			&patient.Address,
			&patient.MedicalInfo,
			&patient.RegisteredBy,
			&patient.CreatedAt,
			&patient.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan patient: %w", err)
		}
		patients = append(patients, patient)
	}

	return patients, rows.Err()
}
