package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// DepartmentHandler handles department CRUD endpoints.
type DepartmentHandler struct {
	service *service.DepartmentService
}

// NewDepartmentHandler creates a new DepartmentHandler.
func NewDepartmentHandler(svc *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: svc}
}

// ListDepartments returns all departments.
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	resp, err := h.service.List(c.Request.Context())
	if err != nil {
		slog.Error("failed to list departments", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list departments"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateDepartment creates a new department.
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	dept, err := h.service.Create(c.Request.Context(), req.Name)
	if err != nil {
		slog.Error("failed to create department", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create department"})
		return
	}

	c.JSON(http.StatusCreated, dept)
}

// UpdateDepartment renames an existing department.
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid department ID"})
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	dept, err := h.service.Update(c.Request.Context(), id, req.Name)
	if err != nil {
		slog.Error("failed to update department", "error", err, "department_id", id)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "department not found"})
		return
	}

	c.JSON(http.StatusOK, dept)
}

// DeleteDepartment removes a department.
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid department ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		slog.Error("failed to delete department", "error", err, "department_id", id)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "department not found"})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "department deleted"})
}
