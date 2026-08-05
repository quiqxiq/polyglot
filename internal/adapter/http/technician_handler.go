package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

type TechnicianHandler struct {
	pgStore *postgres.Store
}

func NewTechnicianHandler(pgStore *postgres.Store) *TechnicianHandler {
	return &TechnicianHandler{pgStore: pgStore}
}

type TechnicianRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Username       string `json:"username" binding:"required"`
	PhoneNumber    string `json:"phone_number" binding:"required"`
	Specialization string `json:"specialization"`
	IsActive       bool   `json:"is_active"`
}

type ToggleActiveRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *TechnicianHandler) ListTechnicians(c *gin.Context) {
	techs, err := h.pgStore.FindAllTechnicians()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"technicians": techs})
}

func (h *TechnicianHandler) CreateTechnician(c *gin.Context) {
	var req TechnicianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tech := &customer.Technician{
		FullName:       req.FullName,
		Username:       req.Username,
		PhoneNumber:    req.PhoneNumber,
		Specialization: req.Specialization,
		IsActive:       true,
	}

	if err := h.pgStore.CreateTechnician(tech); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Teknisi berhasil ditambahkan",
		"technician": tech,
	})
}

func (h *TechnicianHandler) UpdateTechnician(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req TechnicianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tech, err := h.pgStore.FindTechnicianByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teknisi tidak ditemukan"})
		return
	}

	tech.FullName = req.FullName
	tech.Username = req.Username
	tech.PhoneNumber = req.PhoneNumber
	tech.Specialization = req.Specialization

	if err := h.pgStore.UpdateTechnician(tech); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Data teknisi berhasil diperbarui",
		"technician": tech,
	})
}

func (h *TechnicianHandler) ToggleActive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req ToggleActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tech, err := h.pgStore.FindTechnicianByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teknisi tidak ditemukan"})
		return
	}

	tech.IsActive = req.IsActive
	if err := h.pgStore.UpdateTechnician(tech); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Status aktif teknisi berhasil diperbarui",
		"technician": tech,
	})
}

func (h *TechnicianHandler) DeleteTechnician(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.pgStore.DeleteTechnician(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Teknisi berhasil dihapus"})
}
