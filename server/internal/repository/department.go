package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"inventario/shared/models"
)

// DepartmentRepository handles CRUD operations for departments.
type DepartmentRepository struct {
	db *sqlx.DB
}

// NewDepartmentRepository creates a new DepartmentRepository.
func NewDepartmentRepository(db *sqlx.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// List returns all departments ordered by name.
func (r *DepartmentRepository) List(ctx context.Context) ([]models.Department, error) {
	var departments []models.Department
	err := r.db.SelectContext(ctx, &departments, "SELECT * FROM departments ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	if departments == nil {
		departments = []models.Department{}
	}
	return departments, nil
}

// GetByID returns a single department.
func (r *DepartmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Department, error) {
	var dept models.Department
	err := r.db.GetContext(ctx, &dept, "SELECT * FROM departments WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// Create inserts a new department and returns it.
func (r *DepartmentRepository) Create(ctx context.Context, name string) (*models.Department, error) {
	var dept models.Department
	err := r.db.GetContext(ctx, &dept,
		"INSERT INTO departments (id, name, created_at) VALUES (uuid_generate_v4(), $1, NOW()) RETURNING *", name)
	if err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}
	return &dept, nil
}

// Update renames an existing department.
func (r *DepartmentRepository) Update(ctx context.Context, id uuid.UUID, name string) (*models.Department, error) {
	var dept models.Department
	err := r.db.GetContext(ctx, &dept,
		"UPDATE departments SET name = $1 WHERE id = $2 RETURNING *", name, id)
	if err != nil {
		return nil, fmt.Errorf("update department: %w", err)
	}
	return &dept, nil
}

// Delete removes a department. Devices referencing it will have department_id set to NULL (ON DELETE SET NULL).
func (r *DepartmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM departments WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete department: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("department not found")
	}
	return nil
}
