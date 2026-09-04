package repository

import (
	"context"
	"errors"
	"fmt"
	"task-manager-api/internal/models"

	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	conn *pgx.Conn
}

func NewTaskRepository(conn *pgx.Conn) *TaskRepository {
	return &TaskRepository{
		conn: conn,
	}
}

func (r *TaskRepository) CreateTask(title, description, email string, completed bool) (int, error) {
	var id int

	err := r.conn.QueryRow(
		context.Background(),
		`
		INSERT INTO tasks (title, description, user_email, completed) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
		`,
		title, description, email, completed,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *TaskRepository) GetAllByUser(email string) ([]models.UserTasks, error) {

	rows, err := r.conn.Query(
		context.Background(),
		`
		SELECT id, title, description, user_email, completed 
		FROM tasks 
		WHERE user_email = $1
		`, email,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []models.UserTasks

	for rows.Next() {
		var task models.UserTasks
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.UserEmail, &task.TaskCompleted)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) GetByID(email string, id int) (models.UserTasks, error) {
	var task models.UserTasks

	err := r.conn.QueryRow(
		context.Background(),
		`
		SELECT id, title, description, user_email, completed 
		FROM tasks 
		WHERE user_email = $1 AND id = $2
		`,
		email, id,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.UserEmail,
		&task.TaskCompleted,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserTasks{}, models.ErrTaskNotFound
		}
		return models.UserTasks{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) DeleteByID(email string, id int) error {

	commandTag, err := r.conn.Exec(
		context.Background(),
		`
		DELETE FROM tasks 
		WHERE user_email = $1 AND id = $2
		`,
		email, id,
	)

	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return models.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) EditByID(email string, id int, task models.UserTasks) error {

	commandTag, err := r.conn.Exec(
		context.Background(),
		`
		UPDATE tasks SET title = $1, description = $2, completed = $3 
		WHERE user_email = $4 AND id = $5
		`,
		task.Title, task.Description, task.TaskCompleted, email, id,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return models.ErrTaskNotFound
	}

	return nil
}
