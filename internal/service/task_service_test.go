package service

import (
	"context"
	"errors"
	"strings"
	"task-manager-api/internal/models"
	"testing"
)

type fakeTaskRepository struct {
	// Create
	createCalled             bool
	requestedCreateTitle     string
	requestedCreateDesc      string
	requestedCreateUserID    int // ← было Email
	requestedCreateCompleted bool
	returnedCreateID         int

	// Search
	searchCalled          bool
	requestedSearchUserID int // ← было Email
	returnedList          []models.UserTasks

	// SearchTask
	searchTaskCalled    bool
	requestedTaskUserID int // ← было Email
	requestedTaskID     int
	returnedTask        models.UserTasks

	// Delete
	deleteCalled          bool
	requestedDeleteUserID int // ← было Email
	requestedDeleteID     int

	// Update
	updateCalled          bool
	requestedUpdateUserID int // ← было Email
	requestedUpdateID     int
	requestedUpdateTask   models.UserTasks

	returnedError error
}

func (f *fakeTaskRepository) CreateTask(ctx context.Context, title string, description string, userID int, completed bool) (int, error) {
	f.createCalled = true
	f.requestedCreateTitle = title
	f.requestedCreateDesc = description
	f.requestedCreateUserID = userID
	f.requestedCreateCompleted = completed

	return f.returnedCreateID, f.returnedError
}

func (f *fakeTaskRepository) DeleteByID(ctx context.Context, userID int, id int) error {
	f.deleteCalled = true
	f.requestedDeleteUserID = userID
	f.requestedDeleteID = id

	return f.returnedError
}

func (f *fakeTaskRepository) EditByID(ctx context.Context, userID int, id int, task models.UserTasks) error {
	f.updateCalled = true
	f.requestedUpdateUserID = userID
	f.requestedUpdateID = id
	f.requestedUpdateTask = task

	return f.returnedError
}

func (f *fakeTaskRepository) GetAllByUser(ctx context.Context, userID int) ([]models.UserTasks, error) {
	f.searchCalled = true
	f.requestedSearchUserID = userID

	return f.returnedList, f.returnedError
}

func (f *fakeTaskRepository) GetByID(ctx context.Context, userID int, id int) (models.UserTasks, error) {
	f.searchTaskCalled = true
	f.requestedTaskUserID = userID
	f.requestedTaskID = id

	return f.returnedTask, f.returnedError
}

func TestCreate_Success(t *testing.T) {
	fakeTaskRepo := &fakeTaskRepository{returnedCreateID: 42}
	svc := NewTaskUserService(fakeTaskRepo)

	resp, err := svc.Create(context.Background(), models.UserTasks{
		Title:         "testTitle",
		Description:   "testDescription",
		UserID:        42,
		TaskCompleted: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ID != 42 {
		t.Errorf("resp.ID: expected 42, got %d", resp.ID)
	}

	if fakeTaskRepo.requestedCreateTitle != "testTitle" {
		t.Errorf("fakeTaskRepo.requestedCreateTitle: expected %q, got %q", "testTitle", fakeTaskRepo.requestedCreateTitle)
	}
	if fakeTaskRepo.requestedCreateDesc != "testDescription" {
		t.Errorf("fakeTaskRepo.requestedCreateDesc: expected %q, got %q", "testDescription", fakeTaskRepo.requestedCreateDesc)
	}
	if fakeTaskRepo.requestedCreateUserID != 42 {
		t.Errorf("fakeTaskRepo.requestedCreateUserID: expected %q, got %q", 42, fakeTaskRepo.requestedCreateUserID)
	}
	if fakeTaskRepo.requestedCreateCompleted != true {
		t.Errorf("fakeTaskRepo.requestedCreateCompleted: expected true, got %v", fakeTaskRepo.requestedCreateCompleted)
	}

	if !fakeTaskRepo.createCalled {
		t.Error("CreateTask should be called")
	}
}

func TestCreate_Validation(t *testing.T) {
	tests := []struct {
		name        string
		task        models.UserTasks
		expectedErr error
	}{
		{
			name: "empty title",
			task: models.UserTasks{
				Title:       "",
				Description: "desc",
			},
			expectedErr: models.ErrInvalidTitle,
		},
		{
			name: "title too long",
			task: models.UserTasks{
				Title:       strings.Repeat("a", 256),
				Description: "desc",
			},
			expectedErr: models.ErrInvalidTitle,
		},
		{
			name: "empty description",
			task: models.UserTasks{
				Title:       "title",
				Description: "",
			},
			expectedErr: models.ErrInvalidDescription,
		},
		{
			name: "description too long",
			task: models.UserTasks{
				Title:       "title",
				Description: strings.Repeat("a", 1001),
			},
			expectedErr: models.ErrInvalidDescription,
		},
		{
			name: "whitespace title",
			task: models.UserTasks{
				Title:       "   ",
				Description: "desc",
			},
			expectedErr: models.ErrInvalidTitle,
		},
		{
			name: "whitespace description",
			task: models.UserTasks{
				Title:       "title",
				Description: "   ",
			},
			expectedErr: models.ErrInvalidDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRepo := &fakeTaskRepository{}
			svc := NewTaskUserService(fakeRepo)

			_, err := svc.Create(context.Background(), tt.task)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v", tt.expectedErr, err)
			}

			if fakeRepo.createCalled {
				t.Error("CreateTask should not be called for invalid input")
			}
		})
	}
}

func TestCreate_RepositoryError(t *testing.T) {
	dbErr := errors.New("database unavailable")

	fakeTaskRepo := &fakeTaskRepository{
		returnedCreateID: 0,
		returnedError:    dbErr,
	}
	svc := NewTaskUserService(fakeTaskRepo)

	_, err := svc.Create(context.Background(), models.UserTasks{
		Title:         "testTitle",
		Description:   "testDescription",
		UserID:        42,
		TaskCompleted: true,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !fakeTaskRepo.createCalled {
		t.Error("CreateTask was not called, service failed before reaching repository")
	}

	if !errors.Is(err, dbErr) {
		t.Errorf("expected %v, got %v", dbErr, err)
	}
}

func TestSearch_Success(t *testing.T) {
	fakeTaskRepo := &fakeTaskRepository{
		returnedList: []models.UserTasks{
			{
				ID:            42,
				Title:         "testTitle42",
				Description:   "testDescription42",
				UserID:        42,
				TaskCompleted: true,
			},
			{
				ID:            43,
				Title:         "testTitle43",
				Description:   "testDescription43",
				UserID:        42,
				TaskCompleted: true,
			},
			{
				ID:            44,
				Title:         "testTitle44",
				Description:   "testDescription44",
				UserID:        42,
				TaskCompleted: true,
			},
		},
	}
	svc := NewTaskUserService(fakeTaskRepo)

	resp, err := svc.Search(context.Background(), 42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp[0].ID != 42 {
		t.Errorf("resp[0].ID: expected 42, got %d", resp[0].ID)
	}
	if resp[0].Title != "testTitle42" {
		t.Errorf("resp[0].Title: expected %q, got %q", "testTitle42", resp[0].Title)
	}
	if resp[0].Description != "testDescription42" {
		t.Errorf("resp[0].Description: expected %q, got %q", "testDescription42", resp[0].Description)
	}
	if resp[0].UserID != 42 {
		t.Errorf("resp[0].UserID : expected %q, got %q", 42, resp[0].UserID)
	}
	if resp[0].TaskCompleted != true {
		t.Errorf("resp[0].TaskCompleted: expected true, got %v", resp[0].TaskCompleted)
	}

	if fakeTaskRepo.requestedSearchUserID != 42 {
		t.Errorf("fakeTaskRepo.requestedSearchUserID: expected %q, got %q", 42, fakeTaskRepo.requestedSearchUserID)
	}
	if !fakeTaskRepo.searchCalled {
		t.Error("GetAllByUser should be called")
	}
}

func TestSearch_RepositoryError(t *testing.T) {
	dbErr := errors.New("database unavailable")

	fakeTaskRepo := &fakeTaskRepository{returnedError: dbErr}
	svc := NewTaskUserService(fakeTaskRepo)

	_, err := svc.Search(context.Background(), 42)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !fakeTaskRepo.searchCalled {
		t.Error("GetAllByUser was not called, service failed before reaching repository")
	}

	if !errors.Is(err, dbErr) {
		t.Errorf("expected %v, got %v", dbErr, err)
	}
}

func TestSearchTask_Success(t *testing.T) {
	fakeTaskRepo := &fakeTaskRepository{
		returnedTask: models.UserTasks{
			ID:            42,
			Title:         "testTitle42",
			Description:   "testDescription42",
			UserID:        42,
			TaskCompleted: true,
		},
	}
	svc := NewTaskUserService(fakeTaskRepo)

	resp, err := svc.SearchTask(context.Background(), 42, 42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ID != 42 {
		t.Errorf("resp.ID: expected 42, got %d", resp.ID)
	}
	if resp.Title != "testTitle42" {
		t.Errorf("resp.Title: expected %q, got %q", "testTitle42", resp.Title)
	}
	if resp.Description != "testDescription42" {
		t.Errorf("resp.Description: expected %q, got %q", "testDescription42", resp.Description)
	}
	if resp.UserID != 42 {
		t.Errorf("resp.UserID : expected %q, got %q", 42, resp.UserID)
	}
	if resp.TaskCompleted != true {
		t.Errorf("resp.TaskCompleted: expected true, got %v", resp.TaskCompleted)
	}

	if fakeTaskRepo.requestedTaskUserID != 42 {
		t.Errorf("fakeTaskRepo.requestedTaskUserID: expected %q, got %q", 42, fakeTaskRepo.requestedTaskUserID)
	}
	if fakeTaskRepo.requestedTaskID != 42 {
		t.Errorf("fakeTaskRepo.requestedTaskID: expected 42, got %d", fakeTaskRepo.requestedTaskID)
	}
	if !fakeTaskRepo.searchTaskCalled {
		t.Error("GetByID should be called")
	}
}

func TestSearchTask_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
	}{
		{"task not found", models.ErrTaskNotFound},
		{"server error", errors.New("database unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTaskRepo := &fakeTaskRepository{returnedError: tt.serviceErr}
			svc := NewTaskUserService(fakeTaskRepo)

			_, err := svc.SearchTask(context.Background(), 42, 42)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !fakeTaskRepo.searchTaskCalled {
				t.Error("GetByID was not called, service failed before reaching repository")
			}

			if !errors.Is(err, tt.serviceErr) {
				t.Errorf("expected %v, got %v", tt.serviceErr, err)
			}
		})
	}
}

func TestDelete_Success(t *testing.T) {
	fakeTaskRepo := &fakeTaskRepository{}
	svc := NewTaskUserService(fakeTaskRepo)

	err := svc.Delete(context.Background(), 42, 42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fakeTaskRepo.requestedDeleteUserID != 42 {
		t.Errorf("fakeTaskRepo.requestedDeleteUserID: expected %q, got %q", 42, fakeTaskRepo.requestedDeleteUserID)
	}
	if fakeTaskRepo.requestedDeleteID != 42 {
		t.Errorf("fakeTaskRepo.requestedDeleteID: expected 42, got %d", fakeTaskRepo.requestedDeleteID)
	}
	if !fakeTaskRepo.deleteCalled {
		t.Error("DeleteByID should be called")
	}
}

func TestDelete_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
	}{
		{"task not found", models.ErrTaskNotFound},
		{"server error", errors.New("database unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTaskRepo := &fakeTaskRepository{returnedError: tt.serviceErr}
			svc := NewTaskUserService(fakeTaskRepo)

			err := svc.Delete(context.Background(), 42, 42)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !fakeTaskRepo.deleteCalled {
				t.Error("DeleteByID was not called, service failed before reaching repository")
			}

			if !errors.Is(err, tt.serviceErr) {
				t.Errorf("expected %v, got %v", tt.serviceErr, err)
			}
		})
	}
}

func TestUpdate_Success(t *testing.T) {
	fakeTaskRepo := &fakeTaskRepository{}
	svc := NewTaskUserService(fakeTaskRepo)
	task := models.UserTasks{
		Title:         "testTitle42",
		Description:   "testDescription42",
		UserID:        42,
		TaskCompleted: true,
	}

	err := svc.Update(context.Background(), 42, 42, task)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fakeTaskRepo.requestedUpdateUserID != 42 {
		t.Errorf("fakeTaskRepo.requestedUpdateUserID: expected %q, got %q", 42, fakeTaskRepo.requestedUpdateUserID)
	}
	if fakeTaskRepo.requestedUpdateID != 42 {
		t.Errorf("fakeTaskRepo.requestedUpdateID: expected 42, got %d", fakeTaskRepo.requestedUpdateID)
	}
	if fakeTaskRepo.requestedUpdateTask != task {
		t.Errorf("Update got %+v, want %+v", fakeTaskRepo.requestedUpdateTask, task)
	}
	if !fakeTaskRepo.updateCalled {
		t.Error("EditByID should be called")
	}
}

func TestUpdate_Validation(t *testing.T) {
	tests := []struct {
		name        string
		task        models.UserTasks
		expectedErr error
	}{
		{
			name: "empty title",
			task: models.UserTasks{
				Title:       "",
				Description: "desc",
			},
			expectedErr: models.ErrInvalidTitle,
		},
		{
			name: "title too long",
			task: models.UserTasks{
				Title:       strings.Repeat("a", 256),
				Description: "desc",
			},
			expectedErr: models.ErrInvalidTitle,
		},
		{
			name: "empty description",
			task: models.UserTasks{
				Title:       "title",
				Description: "",
			},
			expectedErr: models.ErrInvalidDescription,
		},
		{
			name: "description too long",
			task: models.UserTasks{
				Title:       "title",
				Description: strings.Repeat("a", 1001),
			},
			expectedErr: models.ErrInvalidDescription,
		},
		{
			name: "whitespace title",
			task: models.UserTasks{
				Title:       "   ",
				Description: "desc",
			},
			expectedErr: models.ErrInvalidTitle,
		},
		{
			name: "whitespace description",
			task: models.UserTasks{
				Title:       "title",
				Description: "   ",
			},
			expectedErr: models.ErrInvalidDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRepo := &fakeTaskRepository{}
			svc := NewTaskUserService(fakeRepo)

			err := svc.Update(context.Background(), 42, 42, tt.task)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v", tt.expectedErr, err)
			}

			if fakeRepo.createCalled {
				t.Error("EditByID should not be called for invalid input")
			}
		})
	}
}

func TestUpdate_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
	}{
		{"task not found", models.ErrTaskNotFound},
		{"server error", errors.New("database unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTaskRepo := &fakeTaskRepository{returnedError: tt.serviceErr}
			svc := NewTaskUserService(fakeTaskRepo)

			err := svc.Update(context.Background(), 42, 42, models.UserTasks{
				Title:         "testTitle42",
				Description:   "testDescription42",
				UserID:        42,
				TaskCompleted: true,
			})

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !fakeTaskRepo.updateCalled {
				t.Error("EditByID was not called, service failed before reaching repository")
			}

			if !errors.Is(err, tt.serviceErr) {
				t.Errorf("expected %v, got %v", tt.serviceErr, err)
			}
		})
	}
}
