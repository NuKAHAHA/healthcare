package services

import (
	"context"
	"fmt"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"healthcare-api/internal/repositories"
)

// AuditService handles audit log operations
type AuditService struct {
	auditRepo *repositories.AuditLogRepository
	logger    *logger.Logger
}

// NewAuditService creates a new audit service
func NewAuditService(
	auditRepo *repositories.AuditLogRepository,
	logger *logger.Logger,
) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// GetAuditLogsByPatient retrieves audit logs for a patient (admin only)
// SECURITY: Authorization must be checked in handler
func (s *AuditService) GetAuditLogsByPatient(ctx context.Context, patientID int64, requesterRole string) ([]*models.AuditLog, error) {
	// Only admins can retrieve audit logs
	if requesterRole != models.RoleAdmin {
		s.logger.WarnWithContext("unauthorized audit log access", "audit_log_access", 0, "not admin")
		return nil, fmt.Errorf("not authorized to access audit logs")
	}

	logs, err := s.auditRepo.GetByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
