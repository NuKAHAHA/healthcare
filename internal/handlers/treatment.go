package handlers

import (
	"encoding/json"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/middleware"
	"healthcare-api/internal/models"
	"healthcare-api/internal/services"
	"net/http"
)

// TreatmentHandler handles treatment endpoints
type TreatmentHandler struct {
	treatmentService *services.TreatmentService
	logger           *logger.Logger
	maxBodySize      int64
}

// NewTreatmentHandler creates a new treatment handler
func NewTreatmentHandler(
	treatmentService *services.TreatmentService,
	logger *logger.Logger,
	maxBodySize int64,
) *TreatmentHandler {
	return &TreatmentHandler{
		treatmentService: treatmentService,
		logger:           logger,
		maxBodySize:      maxBodySize,
	}
}

// AddTreatment handles POST /treatments
// SECURITY: Only doctors can add treatments to their own appointments
func (h *TreatmentHandler) AddTreatment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
	defer func() {
		_ = r.Body.Close()
	}()

	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req models.CreateTreatmentRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.logger.WarnWithContext("invalid treatment request", "treatment_add", user.ID, err.Error())
		sendJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	treatment, err := h.treatmentService.AddTreatment(r.Context(), &req, user.ID)
	if err != nil {
		h.logger.WarnWithContext("treatment add failed", "treatment_add", user.ID, err.Error())
		statusCode := http.StatusBadRequest
		message := "Treatment request failed"
		if err.Error() == "not authorized to add treatment to this appointment" {
			statusCode = http.StatusForbidden
			message = "Access denied"
		}
		sendJSONError(w, statusCode, "error", message)
		return
	}

	treatmentDTO := convertTreatmentToDTO(treatment)
	sendJSONSuccess(w, http.StatusCreated, treatmentDTO)
}

// GetReport handles GET /reports/:appointmentId
// SECURITY: Doctors and admins can view reports
func (h *TreatmentHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	// Extract appointment ID from URL
	id, err := extractIDFromPath(r.URL.Path, "/reports/")
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "Invalid appointment ID")
		return
	}

	// Generate report
	report, err := h.treatmentService.GenerateReport(r.Context(), id, user.ID, user.Role)
	if err != nil {
		h.logger.WarnWithContext("report generation failed", "report_generation", user.ID, err.Error())
		statusCode := http.StatusNotFound
		message := "Report not found or appointment has no treatment"
		if err.Error() == "not authorized to access this report" {
			statusCode = http.StatusForbidden
			message = "Access denied"
		}
		sendJSONError(w, statusCode, "error", message)
		return
	}

	reportDTO := convertReportToDTO(report)
	sendJSONSuccess(w, http.StatusOK, reportDTO)
}

// convertTreatmentToDTO converts Treatment model to DTO
func convertTreatmentToDTO(treatment *models.Treatment) models.TreatmentDTO {
	return models.TreatmentDTO{
		ID:            treatment.ID,
		AppointmentID: treatment.AppointmentID,
		PatientID:     treatment.PatientID,
		DoctorID:      treatment.DoctorID,
		Diagnosis:     treatment.Diagnosis,
		Prescription:  treatment.Prescription,
		Notes:         treatment.Notes,
		CreatedAt:     treatment.CreatedAt,
		UpdatedAt:     treatment.UpdatedAt,
	}
}

// convertReportToDTO converts Report model to DTO
func convertReportToDTO(report *models.Report) models.ReportDTO {
	return models.ReportDTO{
		ID:            report.ID,
		AppointmentID: report.AppointmentID,
		PatientID:     report.PatientID,
		DoctorID:      report.DoctorID,
		Diagnosis:     report.Diagnosis,
		Prescription:  report.Prescription,
		Notes:         report.Notes,
		GeneratedAt:   report.GeneratedAt,
	}
}
