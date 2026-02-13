// Package service implements the business logic layer.
package service

import (
	"context"

	"github.com/google/uuid"

	"inventario/server/internal/repository"
	"inventario/shared/dto"
)

// InventoryService orchestrates the inventory submission workflow.
type InventoryService struct {
	inventoryRepo *repository.InventoryRepository
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepo: repo}
}

// ProcessInventory validates and persists a full inventory snapshot for the given device.
func (s *InventoryService) ProcessInventory(ctx context.Context, deviceID uuid.UUID, req *dto.InventoryRequest) error {
	return s.inventoryRepo.Save(ctx, deviceID, req)
}
