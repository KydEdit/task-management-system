package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"task-manager-api/internal/models"
	"task-manager-api/internal/service"
)

type UserHandler struct {
	serviceU *service.UserService
	auth     *service.AuthService
}

type TaskHandler struct {
	serviceT *service.TaskUserService
}

func NewUserHandler(
	s *service.UserService,
	auth *service.AuthService,
) *UserHandler {

	return &UserHandler{
		serviceU: s,
		auth:     auth,
	}
}

func NewTaskHandler(s *service.TaskUserService) *TaskHandler {
	return &TaskHandler{
		serviceT: s,
	}
}

type DeleteResponse struct {
	Success bool `json:"success"`
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdUser, err := h.serviceU.Register(user)
	if err != nil {
		if errors.Is(err, models.ErrInvalidEmail) {
			http.Error(w, "Invalid email", http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrInvalidCompanyID) {
			http.Error(w, "Invalid company", http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrInvalidPassword) {
			http.Error(w, "Invalid password", http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrUserAlreadyExists) {
			http.Error(w, "Duplicate email", http.StatusConflict)
			return
		}
		if errors.Is(err, models.ErrCompanyNotFound) {
			http.Error(w, "Company not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {

	var user models.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	foundUser, err := h.serviceU.Login(user.Email, user.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			http.Error(
				w,
				"Invalid email or password",
				http.StatusUnauthorized,
			)
			return
		}

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	token, err := h.auth.GenerateToken(foundUser.Email)
	if err != nil {
		http.Error(
			w,
			"Could not generate token",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {

	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	meInfo, err := h.serviceU.Me(email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			http.Error(
				w,
				"Invalid email or password",
				http.StatusUnauthorized,
			)
			return
		}

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(meInfo)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {

	var task models.UserTasks

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	task.UserEmail = email

	createdTask, err := h.serviceT.Create(task)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTask)

}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {

	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")

	if idStr == "" {
		tasks, err := h.serviceT.Search(email)
		if err != nil {
			http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
		return
	}

	urlintid, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	task, err := h.serviceT.SearchTask(email, urlintid)
	if err != nil {
		if errors.Is(err, models.ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {

	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	err = h.serviceT.Delete(email, id)
	if err != nil {
		if errors.Is(err, models.ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) EditTask(w http.ResponseWriter, r *http.Request) {
	var task models.UserTasks

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	err = h.serviceT.Update(email, id, task)
	if err != nil {
		if errors.Is(err, models.ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
