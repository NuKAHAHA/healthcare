package handlers

import (
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"healthcare-api/internal/repositories"
	"net/http"
)

type UserHandler struct {
	userRepo *repositories.UserRepository
	logger   *logger.Logger
}

func NewUserHandler(userRepo *repositories.UserRepository, appLogger *logger.Logger) *UserHandler {
	return &UserHandler{userRepo: userRepo, logger: appLogger}
}

// ListUsers handles GET /users?role=doctor
// Admin and registrar only — used by the frontend to populate scheduling dropdowns.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	if role == "" {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "role query parameter is required")
		return
	}

	users, err := h.userRepo.GetByRole(r.Context(), role)
	if err != nil {
		h.logger.WarnWithContext("failed to list users by role", "user_list", 0, err.Error())
		sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch users")
		return
	}

	dtos := make([]models.UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, models.UserDTO{
			ID:        u.ID,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Role:      u.Role,
		})
	}

	sendJSONSuccess(w, http.StatusOK, dtos)
}
