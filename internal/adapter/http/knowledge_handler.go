package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

type KnowledgeHandler struct {
	pgStore *postgres.Store
}

func NewKnowledgeHandler(pgStore *postgres.Store) *KnowledgeHandler {
	return &KnowledgeHandler{
		pgStore: pgStore,
	}
}

type KnowledgeEntryRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Tags    string `json:"tags"`
}

func (h *KnowledgeHandler) ListKnowledge(c *gin.Context) {
	entries, err := h.pgStore.FindAllKnowledgeEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"knowledge_entries": entries})
}

func (h *KnowledgeHandler) CreateKnowledge(c *gin.Context) {
	var req KnowledgeEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry := &knowledge.KnowledgeEntry{
		Title:   req.Title,
		Content: req.Content,
		Tags:    req.Tags,
	}

	if err := h.pgStore.CreateKnowledgeEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Entri basis pengetahuan berhasil dibuat",
		"entry":   entry,
	})
}

func (h *KnowledgeHandler) GetKnowledge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	entry, err := h.pgStore.FindKnowledgeEntryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entri tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entry": entry})
}

func (h *KnowledgeHandler) UpdateKnowledge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req KnowledgeEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.pgStore.FindKnowledgeEntryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entri tidak ditemukan"})
		return
	}

	entry.Title = req.Title
	entry.Content = req.Content
	entry.Tags = req.Tags

	if err := h.pgStore.UpdateKnowledgeEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Entri basis pengetahuan berhasil diperbarui",
		"entry":   entry,
	})
}

func (h *KnowledgeHandler) DeleteKnowledge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.pgStore.DeleteKnowledgeEntry(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entri basis pengetahuan berhasil dihapus"})
}
