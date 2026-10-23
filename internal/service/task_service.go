package service

import (
	"context"
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

func (s *TaskUserService) Create(ctx context.Context, task models.UserTasks) (models.UserTasks, error) {
	if task.Title == "" || len(task.Title) > 255 {
		return models.UserTasks{}, models.ErrInvalidTitle
	}
	if task.Description == "" || len(task.Description) > 1000 {
		return models.UserTasks{}, models.ErrInvalidDescription
	}

	id, err := s.repo.CreateTask(ctx, task.Title, task.Description, task.UserEmail, task.TaskCompleted)
	if err != nil {
		return models.UserTasks{}, err
	}

	task.ID = id
	return task, nil
}

func (s *TaskUserService) Search(ctx context.Context, email string) ([]models.UserTasks, error) {

	tasks, err := s.repo.GetAllByUser(ctx, email)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskUserService) SearchTask(ctx context.Context, email string, id int) (models.UserTasks, error) {

	specificTask, err := s.repo.GetByID(ctx, email, id)
	if err != nil {
		return models.UserTasks{}, err
	}

	return specificTask, nil
}

func (s *TaskUserService) Delete(ctx context.Context, email string, id int) error {
	return s.repo.DeleteByID(ctx, email, id)

}

func (s *TaskUserService) Update(ctx context.Context, email string, id int, task models.UserTasks) error {
	return s.repo.EditByID(ctx, email, id, task)

}
