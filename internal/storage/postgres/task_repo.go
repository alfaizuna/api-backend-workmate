package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskRepository interface {
	Create(ctx context.Context, t *Task) error
	GetByID(ctx context.Context, userID, id string) (*Task, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Task, error)
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, userID, id string) error
}

type taskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) TaskRepository {
	return &taskRepository{pool: pool}
}

func (r *taskRepository) Create(ctx context.Context, t *Task) error {
	const q = `insert into public.tasks (user_id, title, description, status, due_date)
               values ($1, $2, $3, $4, $5)
               returning id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q, t.UserID, t.Title, t.Description, t.Status, t.DueDate).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *taskRepository) GetByID(ctx context.Context, userID, id string) (*Task, error) {
	const q = `select id, user_id, title, description, status, due_date, created_at, updated_at
               from public.tasks where id=$1 and user_id=$2`
	var t Task
	if err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *taskRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]Task, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	const q = `select id, user_id, title, description, status, due_date, created_at, updated_at
               from public.tasks where user_id=$1 order by created_at desc limit $2 offset $3`
	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *taskRepository) Update(ctx context.Context, t *Task) error {
	const q = `update public.tasks set title=$1, description=$2, status=$3, due_date=$4, updated_at=now()
               where id=$5 and user_id=$6 returning updated_at`
	return r.pool.QueryRow(ctx, q, t.Title, t.Description, t.Status, t.DueDate, t.ID, t.UserID).
		Scan(&t.UpdatedAt)
}

func (r *taskRepository) Delete(ctx context.Context, userID, id string) error {
	const q = `delete from public.tasks where id=$1 and user_id=$2`
	_, err := r.pool.Exec(ctx, q, id, userID)
	return err
}
