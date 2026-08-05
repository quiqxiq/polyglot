package http

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

type SessionHandler struct {
	pgStore   *postgres.Store
	waGateway port.WhatsAppGateway
}

func NewSessionHandler(pgStore *postgres.Store, waGateway port.WhatsAppGateway) *SessionHandler {
	return &SessionHandler{
		pgStore:   pgStore,
		waGateway: waGateway,
	}
}

type CreateSessionRequest struct {
	DeviceName  string `json:"device_name" binding:"required"`
	PhoneNumber string `json:"phone_number"`
	WebhookURL  string `json:"webhook_url"`
}

type ToggleBotRequest struct {
	IsBotEnabled bool `json:"is_bot_enabled"`
}

type UpdateWebhookRequest struct {
	WebhookURL string `json:"webhook_url"`
}

func (h *SessionHandler) ListSessions(c *gin.Context) {
	sessions, err := h.pgStore.FindAllSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session := &bot.WASession{
		DeviceName:   req.DeviceName,
		PhoneNumber:  req.PhoneNumber,
		WebhookURL:   req.WebhookURL,
		Status:       bot.StatusNeedsRescan,
		IsBotEnabled: true,
		ConnectedAt:  time.Now(),
	}

	if err := h.pgStore.CreateSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.waGateway.Connect(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memulai koneksi WhatsApp: " + err.Error()})
		return
	}

	qrCode, _ := h.waGateway.GetQRCode(session.ID)

	c.JSON(http.StatusCreated, gin.H{
		"session": session,
		"qr_code": qrCode,
	})
}

func (h *SessionHandler) GetQRCode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	qrCode, err := h.waGateway.GetQRCode(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": id,
		"qr_code":    qrCode,
	})
}

type PairingCodeRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

func (h *SessionHandler) GetPairingCode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req PairingCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor WhatsApp wajib diisi"})
		return
	}

	code, err := h.waGateway.GetPairingCode(uint(id), req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":   id,
		"pairing_code": code,
	})
}

func (h *SessionHandler) ToggleBot(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req ToggleBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.pgStore.FindSessionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sesi tidak ditemukan"})
		return
	}

	session.IsBotEnabled = req.IsBotEnabled
	if err := h.pgStore.UpdateSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status chatbot berhasil diperbarui",
		"session": session,
	})
}

func (h *SessionHandler) UpdateWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.pgStore.FindSessionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sesi tidak ditemukan"})
		return
	}

	session.WebhookURL = req.WebhookURL
	if err := h.pgStore.UpdateSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook URL berhasil diperbarui",
		"session": session,
	})
}

func (h *SessionHandler) LogoutSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	_ = h.waGateway.Logout(uint(id))

	session, err := h.pgStore.FindSessionByID(uint(id))
	if err == nil {
		session.Status = bot.StatusNeedsRescan
		session.PhoneNumber = ""
		session.JID = ""
		_ = h.pgStore.UpdateSession(session)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sesi WhatsApp berhasil di-logout"})
}

func (h *SessionHandler) DeleteSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	_ = h.waGateway.Logout(uint(id))
	if err := h.pgStore.DeleteSession(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sesi berhasil dihapus"})
}

func (h *SessionHandler) ReconnectSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.waGateway.Reconnect(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungkan ulang: " + err.Error()})
		return
	}

	qrCode, _ := h.waGateway.GetQRCode(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "Permintaan hubung ulang WhatsApp berhasil dikirim",
		"qr_code": qrCode,
	})
}

type SendDocumentRequest struct {
	To          string `json:"to" binding:"required"`
	FileName    string `json:"file_name" binding:"required"`
	FileBase64  string `json:"file_base64" binding:"required"`
	ContentType string `json:"content_type"`
	Caption     string `json:"caption"`
}

type SendImageRequest struct {
	To          string `json:"to" binding:"required"`
	ImageBase64 string `json:"image_base64" binding:"required"`
	ContentType string `json:"content_type"`
	Caption     string `json:"caption"`
}

func (h *SessionHandler) SendDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req SendDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileBytes, err := base64.StdEncoding.DecodeString(req.FileBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Base64 file_base64 tidak valid: " + err.Error()})
		return
	}

	if err := h.waGateway.SendDocument(c.Request.Context(), uint(id), req.To, fileBytes, req.FileName, req.ContentType, req.Caption); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim dokumen WA: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dokumen WhatsApp berhasil dikirim"})
}

func (h *SessionHandler) SendImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req SendImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageBytes, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Base64 image_base64 tidak valid: " + err.Error()})
		return
	}

	if err := h.waGateway.SendImage(c.Request.Context(), uint(id), req.To, imageBytes, req.ContentType, req.Caption); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim gambar WA: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gambar WhatsApp berhasil dikirim"})
}

