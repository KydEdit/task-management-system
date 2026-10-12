package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"task-manager-api/internal/models"
	"testing"
)

type fakeTaskService struct {
	// Create
	createCalled  bool
	requestedTask models.UserTasks

	// Search
	searchCalled         bool
	requestedSearchEmail string

	// SearchTask
	searchTaskCalled   bool
	requestedTaskEmail string
	requestedTaskID    int

	// Delete
	deleteCalled         bool
	requestedDeleteEmail string
	requestedDeleteID    int

	// Update
	updateCalled         bool
	requestedUpdateEmail string
	requestedUpdateID    int
	requestedUpdateTask  models.UserTasks

	returnedResponse models.UserTasks
	returnedList     []models.UserTasks
	returnedError    error
}

func (f *fakeTaskService) Create(task models.UserTasks) (models.UserTasks, error) {
	f.createCalled = true
	f.requestedTask = task

	return f.returnedResponse, f.returnedError
}

func (f *fakeTaskService) Delete(email string, id int) error {

	return f.returnedError
}

func (f *fakeTaskService) Search(email string) ([]models.UserTasks, error) {
	f.searchCalled = true
	f.requestedSearchEmail = email

	return f.returnedList, f.returnedError
}

func (f *fakeTaskService) SearchTask(email string, id int) (models.UserTasks, error) {
	f.searchTaskCalled = true
	f.requestedTaskEmail = email
	f.requestedTaskID = id

	return f.returnedResponse, f.returnedError
}

func (f *fakeTaskService) Update(email string, id int, task models.UserTasks) error {

	return f.returnedError
}

func newRequestWithTask(httpMethod, target, email string, body io.Reader) *http.Request {
	req := httptest.NewRequest(httpMethod, target, body)
	ctx := context.WithValue(req.Context(), userContextKey, email)
	return req.WithContext(ctx)
}

func TestCreateTask_Success(t *testing.T) {
	fakeService := &fakeTaskService{
		returnedResponse: models.UserTasks{
			ID:            7,
			Title:         "testTitle",
			Description:   "testDescription",
			TaskCompleted: true,
		},
	}
	h := NewTaskHandler(fakeService)
	rec := httptest.NewRecorder()
	reqBody := models.UserTasks{
		Title:         "testTitle",
		Description:   "testDescription",
		TaskCompleted: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	body := bytes.NewReader(bodyBytes)
	req := newRequestWithTask(http.MethodPost, "/tasks", "test@gmail.com", body)

	h.CreateTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	var resp models.UserTasks
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != 7 {
		t.Errorf("resp.ID: expected 7, got %d", resp.ID)
	}
	if resp.Title != "testTitle" {
		t.Errorf("resp.Title: expected %q, got %q", "testTitle", resp.Title)
	}
	if resp.Description != "testDescription" {
		t.Errorf("resp.Description: expected %q, got %q", "testDescription", resp.Description)
	}

	if fakeService.requestedTask.UserEmail != "test@gmail.com" {
		t.Errorf("Create got UserEmail %q, want %q", fakeService.requestedTask.UserEmail, "test@gmail.com")
	}
	if fakeService.requestedTask.Title != "testTitle" {
		t.Errorf("Create got Title %q, want %q", fakeService.requestedTask.Title, "testTitle")
	}
	if fakeService.requestedTask.Description != "testDescription" {
		t.Errorf("Create got Description %q, want %q", fakeService.requestedTask.Description, "testDescription")
	}
	if resp.TaskCompleted != true {
		t.Errorf("resp.TaskCompleted: expected true, got %v", resp.TaskCompleted)
	}

	if !fakeService.createCalled {
		t.Error("Create should be called")
	}
}

func TestCreateTask_Errors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed json", "{invalid}"},
		{"wrong type", `{"title": 123}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeService := &fakeTaskService{}
			h := NewTaskHandler(fakeService)
			rec := httptest.NewRecorder()
			req := newRequestWithTask(http.MethodPost, "/tasks", "test@gmail.com", strings.NewReader(tt.body))

			h.CreateTask(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rec.Code)
			}

			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("Content-Type: expected application/json, got %q", ct)
			}

			var errResp struct {
				Error string `json:"error"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}
			if errResp.Error == "" {
				t.Error("expected error message in response body")
			}

			if fakeService.createCalled {
				t.Error("Create should not be called for invalid JSON body")
			}
		})
	}
}

func TestCreateTask_Unauthorized(t *testing.T) {
	fakeService := &fakeTaskService{}
	h := NewTaskHandler(fakeService)
	rec := httptest.NewRecorder()
	reqBody := models.UserTasks{
		Title:         "testTitle",
		Description:   "testDescription",
		TaskCompleted: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	body := bytes.NewReader(bodyBytes)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)

	h.CreateTask(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	if fakeService.createCalled {
		t.Error("Create should not be called on unauthorized request")
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response body")
	}
}

func TestCreateTask_ServiceError(t *testing.T) {
	fakeService := &fakeTaskService{
		returnedError: errors.New("database unavailable"),
	}
	h := NewTaskHandler(fakeService)
	rec := httptest.NewRecorder()
	reqBody := models.UserTasks{
		Title:         "testTitle",
		Description:   "testDescription",
		TaskCompleted: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	body := bytes.NewReader(bodyBytes)
	req := newRequestWithTask(http.MethodPost, "/tasks", "test@gmail.com", body)

	h.CreateTask(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	if !fakeService.createCalled {
		t.Error("Create should be called")
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response body")
	}
}

func TestGetTask_Success(t *testing.T) {
	fakeService := &fakeTaskService{
		returnedList: []models.UserTasks{
			{
				ID:            7,
				Title:         "testTitle7",
				Description:   "testDescription7",
				TaskCompleted: true,
			},
			{
				ID:            8,
				Title:         "testTitle8",
				Description:   "testDescription8",
				TaskCompleted: true,
			},
			{
				ID:            9,
				Title:         "testTitle9",
				Description:   "testDescription9",
				TaskCompleted: true,
			},
		},
	}
	h := NewTaskHandler(fakeService)
	rec := httptest.NewRecorder()
	req := newRequestWithTask(http.MethodGet, "/tasks", "test@gmail.com", nil)

	h.GetTask(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	var resp []models.UserTasks
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(resp))
	}
	if resp[0].ID != 7 {
		t.Errorf("resp[0].ID: expected 7, got %d", resp[0].ID)
	}
	if resp[0].Title != "testTitle7" {
		t.Errorf("resp[0].Title: expected %q, got %q", "testTitle7", resp[0].Title)
	}
	if resp[1].ID != 8 {
		t.Errorf("resp[1].ID: expected 8, got %d", resp[1].ID)
	}
	if resp[0].TaskCompleted != true {
		t.Errorf("resp[0].TaskCompleted: expected true, got %v", resp[0].TaskCompleted)
	}

	if fakeService.requestedSearchEmail != "test@gmail.com" {
		t.Errorf("Search got UserEmail %q, want %q", fakeService.requestedSearchEmail, "test@gmail.com")
	}

	if !fakeService.searchCalled {
		t.Error("Search should be called")
	}
}

func TestGetTask_ByID_Success(t *testing.T) {
	fakeService := &fakeTaskService{
		returnedResponse: models.UserTasks{
			ID:            7,
			Title:         "testTitle7",
			Description:   "testDescription7",
			TaskCompleted: true,
		},
	}
	h := NewTaskHandler(fakeService)
	rec := httptest.NewRecorder()
	req := newRequestWithTask(http.MethodGet, "/tasks/{id}", "test@gmail.com", nil)
	req.SetPathValue("id", "7")

	h.GetTask(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	var resp models.UserTasks
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != 7 {
		t.Errorf("resp.ID: expected 7, got %d", resp.ID)
	}
	if resp.Title != "testTitle7" {
		t.Errorf("resp.Title: expected %q, got %q", "testTitle7", resp.Title)
	}
	if resp.Description != "testDescription7" {
		t.Errorf("resp.Description: expected %q, got %q", "testDescription7", resp.Description)
	}
	if resp.TaskCompleted != true {
		t.Errorf("resp.TaskCompleted: expected true, got %v", resp.TaskCompleted)
	}

	if fakeService.requestedTaskEmail != "test@gmail.com" {
		t.Errorf("SearchTask got UserEmail %q, want %q", fakeService.requestedTaskEmail, "test@gmail.com")
	}
	if fakeService.requestedTaskID != 7 {
		t.Errorf("SearchTask got id %d, want %d", fakeService.requestedTaskID, 7)
	}

	if !fakeService.searchTaskCalled {
		t.Error("SearchTask should be called")
	}
}
