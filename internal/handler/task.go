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
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, ok := r.Context().Value(userIDContextKey).(int)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	task.UserID = userID

	createdTask, err := h.serviceT.Create(ctx, task)
	if err != nil {
		if errors.Is(err, models.ErrInvalidTitle) {
			writeError(w, http.StatusBadRequest, "invalid title")
			return
		}
		if errors.Is(err, models.ErrInvalidDescription) {
			writeError(w, http.StatusBadRequest, "invalid description")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, createdTask)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(userIDContextKey).(int)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")

	if idStr == "" {
		tasks, err := h.serviceT.Search(ctx, userID)
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

	task, err := h.serviceT.SearchTask(ctx, userID, urlintid)
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
	ctx := r.Context()

	userID, ok := ctx.Value(userIDContextKey).(int)
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

	err = h.serviceT.Delete(ctx, userID, id)
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
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, ok := ctx.Value(userIDContextKey).(int)
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

	err = h.serviceT.Update(ctx, userID, id, task)
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
