package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"task-manager-api/internal/models"
)

type TaskHandler struct {
	serviceT TaskService
}

func NewTaskHandler(s TaskService) *TaskHandler {
	return &TaskHandler{
		serviceT: s,
	}
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
