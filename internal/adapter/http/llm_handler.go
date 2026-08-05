package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	llmadapter "github.com/quixiq/polyglot/internal/adapter/llm"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/config"
	domainllm "github.com/quixiq/polyglot/internal/domain/llm"
)

type LLMConfigHandler struct {
	pgStore *postgres.Store
	cfg     config.Config
}

func NewLLMConfigHandler(pgStore *postgres.Store, cfg config.Config) *LLMConfigHandler {
	return &LLMConfigHandler{
		pgStore: pgStore,
		cfg:     cfg,
	}
}

type CreateLLMConfigRequest struct {
	Provider        string  `json:"provider" binding:"required"`
	Model           string  `json:"model" binding:"required"`
	APIKey          string  `json:"api_key" binding:"required"`
	Params          string  `json:"params"`
	MaxOutputTokens int     `json:"max_output_tokens"`
	CostPer1MInput  float64 `json:"cost_per_1m_input"`
	CostPer1MOutput float64 `json:"cost_per_1m_output"`
}

type UpdateLLMConfigRequest struct {
	Provider        string  `json:"provider"`
	Model           string  `json:"model"`
	APIKey          string  `json:"api_key"`
	Params          string  `json:"params"`
	MaxOutputTokens int     `json:"max_output_tokens"`
	CostPer1MInput  float64 `json:"cost_per_1m_input"`
	CostPer1MOutput float64 `json:"cost_per_1m_output"`
}

func (h *LLMConfigHandler) ListConfigs(c *gin.Context) {
	configs, err := h.pgStore.FindAllLLMConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func (h *LLMConfigHandler) CreateConfig(c *gin.Context) {
	var req CreateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedKey, err := config.Encrypt(req.APIKey, h.cfg.EncryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi API Key"})
		return
	}

	maxTokens := req.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = h.cfg.LLMMaxOutputTokens
	}

	costIn := req.CostPer1MInput
	costOut := req.CostPer1MOutput
	if costIn <= 0 && costOut <= 0 {
		costIn, costOut = domainllm.GetDefaultModelPricing(req.Provider, req.Model)
	}

	llmCfg := &domainllm.LLMConfig{
		Provider:        req.Provider,
		Model:           req.Model,
		APIKeyEncrypted: encryptedKey,
		Params:          req.Params,
		MaxOutputTokens: maxTokens,
		CostPer1MInput:  costIn,
		CostPer1MOutput: costOut,
		IsActive:        false,
	}

	if err := h.pgStore.CreateLLMConfig(llmCfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.pgStore.PopulateLLMConfigAnalytics(llmCfg)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Konfigurasi LLM berhasil dibuat",
		"config":  llmCfg,
	})
}

func (h *LLMConfigHandler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	llmCfg, err := h.pgStore.FindLLMConfigByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Konfigurasi tidak ditemukan"})
		return
	}

	if req.Provider != "" {
		llmCfg.Provider = req.Provider
	}
	if req.Model != "" {
		llmCfg.Model = req.Model
	}
	if req.Params != "" {
		llmCfg.Params = req.Params
	}
	if req.MaxOutputTokens > 0 {
		llmCfg.MaxOutputTokens = req.MaxOutputTokens
	}
	if req.CostPer1MInput > 0 {
		llmCfg.CostPer1MInput = req.CostPer1MInput
	}
	if req.CostPer1MOutput > 0 {
		llmCfg.CostPer1MOutput = req.CostPer1MOutput
	}
	if req.APIKey != "" {
		encryptedKey, err := config.Encrypt(req.APIKey, h.cfg.EncryptionKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi API Key"})
			return
		}
		llmCfg.APIKeyEncrypted = encryptedKey
	}

	if err := h.pgStore.UpdateLLMConfig(llmCfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.pgStore.PopulateLLMConfigAnalytics(llmCfg)

	c.JSON(http.StatusOK, gin.H{
		"message": "Konfigurasi LLM berhasil diperbarui",
		"config":  llmCfg,
	})
}

func (h *LLMConfigHandler) ActivateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.pgStore.SetActiveLLMConfig(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Konfigurasi LLM berhasil diaktifkan"})
}

func (h *LLMConfigHandler) TestConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	llmCfg, err := h.pgStore.FindLLMConfigByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Konfigurasi tidak ditemukan"})
		return
	}

	provider, err := llmadapter.NewProvider(llmCfg, h.cfg.EncryptionKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal menginisialisasi provider: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := provider.TestConnection(ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed",
			"error":  "Tes koneksi gagal: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Tes koneksi berhasil! Provider dan API Key valid.",
	})
}

func (h *LLMConfigHandler) DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.pgStore.DeleteLLMConfig(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Konfigurasi LLM berhasil dihapus"})
}
