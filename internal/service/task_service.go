package service

import (
	"task-manager-api/internal/models"
	"task-manager-api/internal/repository"
)

type TaskUserService struct {
	repo *repository.TaskRepository
}

func NewTaskUserService(repo *repository.TaskRepository) *TaskUserService {
	return &TaskUserService{
		repo: repo,
	}
}

func (s *TaskUserService) Create(task models.UserTasks) (models.UserTasks, error) {

	id, err := s.repo.CreateTask(task.Title, task.Description, task.UserEmail, task.TaskCompleted)
	if err != nil {
		return models.UserTasks{}, err
	}

	task.ID = id
	return task, nil
}

func (s *TaskUserService) Search(email string) ([]models.UserTasks, error) {

	tasks, err := s.repo.GetAllByUser(email)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskUserService) SearchTask(email string, id int) (models.UserTasks, error) {

	specificTask, err := s.repo.GetByID(email, id)
	if err != nil {
		return models.UserTasks{}, err
	}

	return specificTask, nil
}

func (s *TaskUserService) Delete(email string, id int) error {
	return s.repo.DeleteByID(email, id)

}

func (s *TaskUserService) Update(email string, id int, task models.UserTasks) error {
	return s.repo.EditByID(email, id, task)

}
