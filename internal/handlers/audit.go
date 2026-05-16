package handlers

import (
	"healthcare-api/internal/logger"
	"healthcare-api/internal/middleware"
	"healthcare-api/internal/models"
	"healthcare-api/internal/services"
	"net/http"
)

// AuditHandler handles audit log endpoints
type AuditHandler struct {
	auditService *services.AuditService
	logger       *logger.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService, logger *logger.Logger) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// GetAuditLogs handles GET /audit-logs/:patientId
// SECURITY: Admin only - critical security endpoint
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	// Extract patient ID from URL
	id, err := extractIDFromPath(r.URL.Path, "/audit-logs/")
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "Invalid patient ID")
		return
	}

	// Get audit logs (service checks authorization)
	logs, err := h.auditService.GetAuditLogsByPatient(r.Context(), id, user.Role)
	if err != nil {
		h.logger.WarnWithContext("audit log access denied", "audit_log_access", user.ID, err.Error())
		sendJSONError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	// Convert to DTOs
	var logDTOs []models.AuditLogDTO
	for _, log := range logs {
		logDTOs = append(logDTOs, models.AuditLogDTO{
			ID:         log.ID,
			UserID:     log.UserID,
			Action:     log.Action,
			Resource:   log.Resource,
			ResourceID: log.ResourceID,
			Status:     log.Status,
			Details:    log.Details,
			CreatedAt:  log.CreatedAt,
		})
	}

	sendJSONSuccess(w, http.StatusOK, logDTOs)
}
