package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"inventario/server/internal/repository"
	"inventario/shared/dto"
	"inventario/shared/models"
)

// DepartmentService handles department business logic.
type DepartmentService struct {
	deptRepo *repository.DepartmentRepository
}

// NewDepartmentService creates a new DepartmentService.
func NewDepartmentService(repo *repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{deptRepo: repo}
}

// List returns all departments.
func (s *DepartmentService) List(ctx context.Context) (*dto.DepartmentListResponse, error) {
	departments, err := s.deptRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	return &dto.DepartmentListResponse{
		Departments: departments,
		Total:       len(departments),
	}, nil
}

// Create adds a new department.
func (s *DepartmentService) Create(ctx context.Context, name string) (*models.Department, error) {
	dept, err := s.deptRepo.Create(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}
	return dept, nil
}

// Update renames a department.
func (s *DepartmentService) Update(ctx context.Context, id uuid.UUID, name string) (*models.Department, error) {
	dept, err := s.deptRepo.Update(ctx, id, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("update department: %w", err)
	}
	return dept, nil
}

// Delete removes a department.
func (s *DepartmentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deptRepo.Delete(ctx, id)
}
