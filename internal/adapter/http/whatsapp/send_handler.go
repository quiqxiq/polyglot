package whatsapp

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/quixiq/polyglot/pkg/response"
)

func (h *Handler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	to := strings.TrimSpace(req.To)
	if to == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "to is required")
		return
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "content is required")
		return
	}

	if h.waGateway == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "WhatsApp gateway unavailable")
		return
	}

	if err := h.waGateway.SendMessage(req.SessionID, to, content); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "message queued for delivery",
	})
}

func (h *Handler) handleSendImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.waGateway == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "WhatsApp gateway unavailable")
		return
	}

	var sessionID uint
	var to, caption, contentType string
	var imageBytes []byte

	contentTypeHeader := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentTypeHeader, "multipart/form-data") {
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Failed to parse multipart form")
			return
		}
		sID, _ := strconv.ParseUint(r.FormValue("session_id"), 10, 32)
		sessionID = uint(sID)
		to = strings.TrimSpace(r.FormValue("to"))
		caption = r.FormValue("caption")

		file, header, err := r.FormFile("file")
		if err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Missing image file in multipart form")
			return
		}
		defer func() { _ = file.Close() }()

		contentType = header.Header.Get("Content-Type")
		imageBytes, err = io.ReadAll(file)
		if err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Failed to read image file data")
			return
		}
	} else {
		var req SendImageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}
		sessionID = req.SessionID
		to = strings.TrimSpace(req.To)
		caption = req.Caption
		contentType = req.ContentType

		rawBytes, err := base64.StdEncoding.DecodeString(req.ImageBase64)
		if err != nil || len(rawBytes) == 0 {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "image_base64 must be valid non-empty base64 string")
			return
		}
		imageBytes = rawBytes
	}

	if sessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}
	if to == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "to is required")
		return
	}

	if err := h.waGateway.SendImage(r.Context(), sessionID, to, imageBytes, contentType, caption); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "image queued for delivery",
	})
}

func (h *Handler) handleSendDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.waGateway == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "WhatsApp gateway unavailable")
		return
	}

	var sessionID uint
	var to, caption, fileName, contentType string
	var fileBytes []byte

	contentTypeHeader := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentTypeHeader, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Failed to parse multipart form")
			return
		}
		sID, _ := strconv.ParseUint(r.FormValue("session_id"), 10, 32)
		sessionID = uint(sID)
		to = strings.TrimSpace(r.FormValue("to"))
		caption = r.FormValue("caption")

		file, header, err := r.FormFile("file")
		if err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Missing document file in multipart form")
			return
		}
		defer func() { _ = file.Close() }()

		fileName = header.Filename
		contentType = header.Header.Get("Content-Type")
		fileBytes, err = io.ReadAll(file)
		if err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Failed to read document file data")
			return
		}
	} else {
		var req SendDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}
		sessionID = req.SessionID
		to = strings.TrimSpace(req.To)
		caption = req.Caption
		fileName = req.FileName
		contentType = req.ContentType

		rawBytes, err := base64.StdEncoding.DecodeString(req.DocumentBase64)
		if err != nil || len(rawBytes) == 0 {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "document_base64 must be valid non-empty base64 string")
			return
		}
		fileBytes = rawBytes
	}

	if sessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}
	if to == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "to is required")
		return
	}
	if fileName == "" {
		fileName = "document.bin"
	}

	if err := h.waGateway.SendDocument(r.Context(), sessionID, to, fileBytes, fileName, contentType, caption); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "document queued for delivery",
	})
}

func (h *Handler) handleChatPresence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req ChatPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	to := strings.TrimSpace(req.To)
	if to == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "to is required")
		return
	}

	if h.waGateway == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "WhatsApp gateway unavailable")
		return
	}

	if err := h.waGateway.SendChatPresence(r.Context(), req.SessionID, to, req.State, req.Media); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "presence updated",
	})
}

func (h *Handler) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req MarkReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	if len(req.MessageIDs) == 0 {
		response.WriteHTTPError(w, domainWA.ErrEmptyMessageID)
		return
	}

	chatJID := strings.TrimSpace(req.ChatJID)
	if chatJID == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "chat_jid is required")
		return
	}

	if h.waGateway == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "WhatsApp gateway unavailable")
		return
	}

	if err := h.waGateway.MarkRead(r.Context(), req.SessionID, req.MessageIDs, chatJID, req.SenderJID); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "messages marked as read",
	})
}
