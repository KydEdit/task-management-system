package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"task-manager-api/internal/models"
)

type UserHandler struct {
	serviceU UserService
	auth     AuthService
}

type TaskHandler struct {
	serviceT TaskService
}

func NewUserHandler(
	s UserService,
	auth AuthService,
) *UserHandler {

	return &UserHandler{
		serviceU: s,
		auth:     auth,
	}
}

func NewTaskHandler(s TaskService) *TaskHandler {
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
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdUser, err := h.serviceU.Register(user)
	if err != nil {
		if errors.Is(err, models.ErrInvalidEmail) {
			writeError(w, http.StatusBadRequest, "invalid email")
			return
		}
		if errors.Is(err, models.ErrInvalidCompanyID) {
			writeError(w, http.StatusBadRequest, "invalid company")
			return
		}
		if errors.Is(err, models.ErrInvalidPassword) {
			writeError(w, http.StatusBadRequest, "invalid password")
			return
		}
		if errors.Is(err, models.ErrUserAlreadyExists) {
			writeError(w, http.StatusConflict, "duplicate email")
			return
		}
		if errors.Is(err, models.ErrCompanyNotFound) {
			writeError(w, http.StatusNotFound, "company not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	writeJSON(w, http.StatusCreated, createdUser)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var user models.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	foundUser, err := h.serviceU.Login(user.Email, user.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	token, err := h.auth.GenerateToken(foundUser.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	meInfo, err := h.serviceU.Me(email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "user no longer exists")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, meInfo)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.UserTasks

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	task.UserEmail = email

	createdTask, err := h.serviceT.Create(task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, createdTask)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")

	if idStr == "" {
		tasks, err := h.serviceT.Search(email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get tasks")
			return
		}

		writeJSON(w, http.StatusOK, tasks)
		return
	}

	urlintid, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	task, err := h.serviceT.SearchTask(email, urlintid)
	if err != nil {
		if errors.Is(err, models.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	err = h.serviceT.Delete(email, id)
	if err != nil {
		if errors.Is(err, models.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) EditTask(w http.ResponseWriter, r *http.Request) {
	var task models.UserTasks

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	err = h.serviceT.Update(email, id, task)
	if err != nil {
		if errors.Is(err, models.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
