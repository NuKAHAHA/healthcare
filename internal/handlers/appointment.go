package handlers

import (
	"encoding/json"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/middleware"
	"healthcare-api/internal/models"
	"healthcare-api/internal/repositories"
	"healthcare-api/internal/services"
	"net/http"
	"strconv"
)

type AppointmentHandler struct {
	appointmentService *services.AppointmentService
	appointmentRepo    *repositories.AppointmentRepository
	logger             *logger.Logger
	maxBodySize        int64
}

func NewAppointmentHandler(
	appointmentService *services.AppointmentService,
	appointmentRepo *repositories.AppointmentRepository,
	appLogger *logger.Logger,
	maxBodySize int64,
) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: appointmentService,
		appointmentRepo:    appointmentRepo,
		logger:             appLogger,
		maxBodySize:        maxBodySize,
	}
}

// CreateAppointment handles POST /appointments
// @Summary      Create appointment
// @Description  Schedule a new patient appointment. Registrar and Admin only.
// @Tags         Appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CreateAppointmentRequest  true  "Appointment data"
// @Success      201   {object}  models.AppointmentDTO
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Router       /appointments [post]
func (h *AppointmentHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
	defer func() { _ = r.Body.Close() }()

	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req models.CreateAppointmentRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		h.logger.WarnWithContext("invalid appointment request", "appointment_creation", user.ID, err.Error())
		sendJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	appointment, err := h.appointmentService.CreateAppointment(r.Context(), &req, user.ID)
	if err != nil {
		h.logger.WarnWithContext("appointment creation failed", "appointment_creation", user.ID, err.Error())
		sendJSONError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	sendJSONSuccess(w, http.StatusCreated, convertAppointmentToDTO(appointment))
}

// ListAppointments handles GET /appointments
// @Summary      List appointments
// @Description  Retrieve appointments. Doctors see only their own; admins and registrars see all.
// @Tags         Appointments
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int  false  "Page (default 1)"
// @Param        pageSize  query  int  false  "Page size (default 50)"
// @Success      200  {array}   models.AppointmentDTO
// @Failure      401  {object}  models.ErrorResponse
// @Router       /appointments [get]
func (h *AppointmentHandler) ListAppointments(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 50)

	var appointments []*models.Appointment
	switch user.Role {
	case "doctor":
		appointments, err = h.appointmentRepo.GetByDoctorID(r.Context(), user.ID, page, pageSize)
	default:
		// admin and registrar see all
		appointments, err = h.appointmentRepo.GetAll(r.Context(), page, pageSize)
	}
	if err != nil {
		h.logger.ErrorWithContext("failed to list appointments", "appointment_list", user.ID, err.Error())
		sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch appointments")
		return
	}

	dtos := make([]models.AppointmentDTO, 0, len(appointments))
	for _, a := range appointments {
		dtos = append(dtos, convertAppointmentToDTO(a))
	}
	sendJSONSuccess(w, http.StatusOK, dtos)
}

// UpdateStatus handles PATCH /appointments/{id}/status
func (h *AppointmentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
	defer func() { _ = r.Body.Close() }()

	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	id, parseErr := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if parseErr != nil || id <= 0 {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "Invalid appointment ID")
		return
	}

	var req models.UpdateAppointmentStatusRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	if err := h.appointmentService.UpdateStatus(r.Context(), id, req.Status, user.ID); err != nil {
		h.logger.WarnWithContext("appointment status update failed", "appointment_status_update", user.ID, err.Error())
		sendJSONError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	sendJSONSuccess(w, http.StatusOK, map[string]string{"status": req.Status})
}

func convertAppointmentToDTO(a *models.Appointment) models.AppointmentDTO {
	return models.AppointmentDTO{
		ID:          a.ID,
		PatientID:   a.PatientID,
		PatientName: a.PatientName,
		DoctorID:    a.DoctorID,
		ScheduledAt: a.ScheduledAt,
		Reason:      a.Reason,
		Status:      a.Status,
		Notes:       a.Notes,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
