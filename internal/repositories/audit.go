package repositories

import (
	"context"
	"fmt"
	"healthcare-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditLogRepository handles audit log database operations
type AuditLogRepository struct {
	db *pgxpool.Pool
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry using parameterized query
// CRITICAL SECURITY: All audit operations must use parameterized queries
func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	// SECURITY: Using parameterized query to prevent SQL injection
	query := `
		INSERT INTO audit_logs (user_id, action, resource, resource_id, status, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, user_id, action, resource, resource_id, status, details, created_at
	`

	err := r.db.QueryRow(ctx, query,
		log.UserID,
		log.Action,
		log.Resource,
		log.ResourceID,
		log.Status,
		log.Details,
	).Scan(
		&log.ID,
		&log.UserID,
		&log.Action,
		&log.Resource,
		&log.ResourceID,
		&log.Status,
		&log.Details,
		&log.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return log, nil
}

// GetByPatientID retrieves audit logs for a patient (admin only) using parameterized query
func (r *AuditLogRepository) GetByPatientID(ctx context.Context, patientID int64) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog

	// SECURITY: Using parameterized query to prevent SQL injection
	query := `
		SELECT id, user_id, action, resource, resource_id, status, details, created_at
		FROM audit_logs WHERE resource_id = $1 AND resource = 'patient' ORDER BY created_at DESC LIMIT 100
	`

	rows, err := r.db.Query(ctx, query, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch audit logs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		log := &models.AuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.ResourceID,
			&log.Status,
			&log.Details,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
