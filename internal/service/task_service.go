package service

import (
	"context"
	"strings"
	"task-manager-api/internal/models"
)

type TaskUserService struct {
	repo TaskRepository
}

func NewTaskUserService(repo TaskRepository) *TaskUserService {
	return &TaskUserService{
		repo: repo,
	}
}

func validateTask(task *models.UserTasks) error {
	title := strings.TrimSpace(task.Title)
	if title == "" || len(title) > 255 {
		return models.ErrInvalidTitle
	}
	desc := strings.TrimSpace(task.Description)
	if desc == "" || len(desc) > 1000 {
		return models.ErrInvalidDescription
	}
	task.Title = title
	task.Description = desc
	return nil
}

func (s *TaskUserService) Create(ctx context.Context, task models.UserTasks) (models.UserTasks, error) {
	if err := validateTask(&task); err != nil {
		return models.UserTasks{}, err
	}

	id, err := s.repo.CreateTask(ctx, task.Title, task.Description, task.UserID, task.TaskCompleted)
	if err != nil {
		return models.UserTasks{}, err
	}

	task.ID = id
	return task, nil
}

func (s *TaskUserService) Search(ctx context.Context, userID int) ([]models.UserTasks, error) {

	tasks, err := s.repo.GetAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskUserService) SearchTask(ctx context.Context, userID int, id int) (models.UserTasks, error) {

	specificTask, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return models.UserTasks{}, err
	}

	return specificTask, nil
}

func (s *TaskUserService) Delete(ctx context.Context, userID int, id int) error {
	return s.repo.DeleteByID(ctx, userID, id)

}

func (s *TaskUserService) Update(ctx context.Context, userID int, id int, task models.UserTasks) error {
	if err := validateTask(&task); err != nil {
		return err
	}

	return s.repo.EditByID(ctx, userID, id, task)

}
