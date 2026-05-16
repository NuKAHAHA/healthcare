package handlers

import (
	"encoding/json"
	"fmt"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/middleware"
	"healthcare-api/internal/models"
	"healthcare-api/internal/repositories"
	"healthcare-api/internal/services"
	"net/http"
	"strconv"
	"strings"
)

type PatientHandler struct {
	patientService  *services.PatientService
	appointmentRepo *repositories.AppointmentRepository
	userRepo        *repositories.UserRepository
	logger          *logger.Logger
	maxBodySize     int64
}

func NewPatientHandler(
	patientService *services.PatientService,
	appointmentRepo *repositories.AppointmentRepository,
	userRepo *repositories.UserRepository,
	appLogger *logger.Logger,
	maxBodySize int64,
) *PatientHandler {
	return &PatientHandler{
		patientService:  patientService,
		appointmentRepo: appointmentRepo,
		userRepo:        userRepo,
		logger:          appLogger,
		maxBodySize:     maxBodySize,
	}
}

// RegisterPatient handles POST /patients
// @Summary      Register patient
// @Description  Create a new patient record. Registrar and Admin only.
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CreatePatientRequest  true  "Patient data"
// @Success      201   {object}  models.PatientDTO
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Router       /patients [post]
func (h *PatientHandler) RegisterPatient(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
	defer func() { _ = r.Body.Close() }()

	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req models.CreatePatientRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		h.logger.WarnWithContext("invalid patient registration request", "patient_registration", user.ID, err.Error())
		sendJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	patient, err := h.patientService.RegisterPatient(r.Context(), &req, user.ID)
	if err != nil {
		h.logger.WarnWithContext("patient registration failed", "patient_registration", user.ID, err.Error())
		sendJSONError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	sendJSONSuccess(w, http.StatusCreated, convertPatientToDTO(patient))
}

// ListPatients handles GET /patients
// @Summary      List patients
// @Description  Retrieve paginated list of patients. Registrar and Admin only.
// @Tags         Patients
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int  false  "Page number (default 1)"
// @Param        pageSize  query  int  false  "Items per page (default 50, max 100)"
// @Success      200  {array}   models.PatientDTO
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Router       /patients [get]
func (h *PatientHandler) ListPatients(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 50)

	var patients []*models.Patient
	switch user.Role {
	case "doctor":
		patients, err = h.patientService.GetPatientsByDoctor(r.Context(), user.ID, page, pageSize)
	default:
		patients, err = h.patientService.GetAllPatients(r.Context(), page, pageSize)
	}
	if err != nil {
		h.logger.ErrorWithContext("failed to list patients", "patient_list", user.ID, err.Error())
		sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch patients")
		return
	}

	dtos := make([]models.PatientDTO, 0, len(patients))
	for _, p := range patients {
		dtos = append(dtos, convertPatientToDTO(p))
	}
	sendJSONSuccess(w, http.StatusOK, dtos)
}

// GetPatient handles GET /patients/:id
// @Summary      Get patient
// @Description  Get patient by ID. Doctors can only access patients they have appointments with.
// @Tags         Patients
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Patient ID"
// @Success      200  {object}  models.PatientDTO
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /patients/{id} [get]
func (h *PatientHandler) GetPatient(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	id, err := extractIDFromPath(r.URL.Path, "/patients/")
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "Invalid patient ID")
		return
	}

	patient, err := h.patientService.GetPatient(r.Context(), id)
	if err != nil {
		h.logger.WarnWithContext("patient not found", "patient_access", user.ID,
			fmt.Sprintf("patient_id=%d", id))
		sendJSONError(w, http.StatusNotFound, "not_found", "Patient not found")
		return
	}

	// Object-level authorization: doctors can only view patients they have appointments with
	if user.Role == "doctor" {
		hasAccess, err := h.appointmentRepo.HasAppointmentWithPatient(r.Context(), user.ID, id)
		if err != nil {
			h.logger.ErrorWithContext("appointment check failed", "patient_access", user.ID, err.Error())
			sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to verify access")
			return
		}
		if !hasAccess {
			h.logger.WarnWithContext("doctor accessed unauthorized patient", "patient_access", user.ID,
				fmt.Sprintf("patient_id=%d ip=%s", id, middleware.ClientIP(r)))
			sendJSONError(w, http.StatusForbidden, "forbidden", "Access denied")
			return
		}
	}

	// Audit log every patient record access
	h.patientService.LogPatientAccess(r.Context(), user.ID, id, middleware.ClientIP(r))

	sendJSONSuccess(w, http.StatusOK, convertPatientToDTO(patient))
}

// ── helpers ──────────────────────────────────────────────────────────────────

func extractIDFromPath(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	// Strip any trailing slash or query string
	if idx := strings.IndexAny(idStr, "/?"); idx != -1 {
		idStr = idStr[:idx]
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}

func convertPatientToDTO(patient *models.Patient) models.PatientDTO {
	return models.PatientDTO{
		ID:           patient.ID,
		FirstName:    patient.FirstName,
		LastName:     patient.LastName,
		Email:        patient.Email,
		Phone:        patient.Phone,
		DateOfBirth:  patient.DateOfBirth,
		Gender:       patient.Gender,
		Address:      patient.Address,
		MedicalInfo:  patient.MedicalInfo,
		RegisteredBy: patient.RegisteredBy,
		CreatedAt:    patient.CreatedAt,
		UpdatedAt:    patient.UpdatedAt,
	}
}
