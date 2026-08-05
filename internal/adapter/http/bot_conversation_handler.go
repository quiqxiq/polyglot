package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

type ConversationHandler struct {
	convService *business.ConversationService
	waGateway   port.WhatsAppGateway
	sseHub      *ws.SSEHub
}

func NewConversationHandler(convService *business.ConversationService, waGateway port.WhatsAppGateway, sseHub *ws.SSEHub) *ConversationHandler {
	return &ConversationHandler{
		convService: convService,
		waGateway:   waGateway,
		sseHub:      sseHub,
	}
}

type SendAgentMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *ConversationHandler) ListConversations(c *gin.Context) {
	status := c.Query("status")
	convs, err := h.convService.ListConversations(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": convs})
}

func (h *ConversationHandler) GetConversation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	conv, err := h.convService.GetConversationWithMessages(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Percakapan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversation": conv})
}

func (h *ConversationHandler) TakeOver(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	var userID uint
	if exists {
		if uid, ok := userIDVal.(uint); ok {
			userID = uid
		}
	}

	if err := h.convService.TakeOver(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	conv, _ := h.convService.GetConversation(uint(id))
	if h.sseHub != nil && conv != nil {
		h.sseHub.Broadcast("conversation_update", conv)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Percakapan berhasil diambil alih. Bot otomatis dihentikan untuk obrolan ini.",
		"conversation": conv,
	})
}

func (h *ConversationHandler) ResetBot(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.convService.ResetBot(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	conv, _ := h.convService.GetConversation(uint(id))
	if h.sseHub != nil && conv != nil {
		h.sseHub.Broadcast("conversation_update", conv)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Bot AI berhasil diaktifkan kembali untuk percakapan ini.",
		"conversation": conv,
	})
}

func (h *ConversationHandler) SendMessage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req SendAgentMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := h.convService.GetConversation(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Percakapan tidak ditemukan"})
		return
	}

	if err := h.waGateway.SendMessage(conv.SessionID, conv.CustomerWANumber, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim pesan WhatsApp: " + err.Error()})
		return
	}

	msg, err := h.convService.AddMessage(conv.ID, "agent", req.Content, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if h.sseHub != nil {
		h.sseHub.Broadcast("new_message", msg)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pesan berhasil dikirim",
		"data":    msg,
	})
}

func (h *ConversationHandler) CloseConversation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.convService.CloseConversation(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	conv, _ := h.convService.GetConversation(uint(id))
	if h.sseHub != nil && conv != nil {
		h.sseHub.Broadcast("conversation_update", conv)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Percakapan selesai"})
}
