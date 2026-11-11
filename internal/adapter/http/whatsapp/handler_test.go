package whatsapp_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	httpWA "github.com/quixiq/polyglot/internal/adapter/http/whatsapp"
	"github.com/quixiq/polyglot/internal/domain/customer"
	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
	ucWA "github.com/quixiq/polyglot/internal/usecase/whatsapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSessionRepo struct {
	sessions []domainWA.WASession
}

func (m *mockSessionRepo) FindAllSessions(context.Context) ([]domainWA.WASession, error) {
	return m.sessions, nil
}
func (m *mockSessionRepo) FindSessionByID(_ context.Context, id uint) (*domainWA.WASession, error) {
	for _, s := range m.sessions {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, domainWA.ErrNotFound
}
func (m *mockSessionRepo) CreateSession(_ context.Context, s *domainWA.WASession) error {
	s.ID = uint(len(m.sessions) + 1)
	m.sessions = append(m.sessions, *s)
	return nil
}
func (m *mockSessionRepo) UpdateSession(context.Context, *domainWA.WASession) error { return nil }
func (m *mockSessionRepo) DeleteSession(context.Context, uint) error                { return nil }
func (m *mockSessionRepo) PurgeSession(context.Context, uint) error                 { return nil }

type mockUserRepo struct {
	users map[string]*customer.User
}

func (m *mockUserRepo) Create(context.Context, *customer.User) error           { return nil }
func (m *mockUserRepo) FindByID(context.Context, uint) (*customer.User, error) { return nil, nil }
func (m *mockUserRepo) FindByUsername(context.Context, string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByEmail(context.Context, string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByPhone(_ context.Context, phone string) (*customer.User, error) {
	if u, ok := m.users[phone]; ok {
		return u, nil
	}
	return nil, nil
}
func (m *mockUserRepo) FindAllByPhone(_ context.Context, phone string) ([]*customer.User, error) {
	if u, ok := m.users[phone]; ok {
		return []*customer.User{u}, nil
	}
	return nil, nil
}
func (m *mockUserRepo) Count(context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepo) List(context.Context, int, int, string) ([]*customer.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) FindAll(context.Context) ([]*customer.User, error) { return nil, nil }
func (m *mockUserRepo) FindByRoles(context.Context, []string, bool) ([]*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(context.Context, *customer.User) error { return nil }
func (m *mockUserRepo) Delete(context.Context, uint) error           { return nil }
func (m *mockUserRepo) UpdatePassword(context.Context, uint, string) error {
	return nil
}
func (m *mockUserRepo) UpdateStatus(context.Context, uint, bool) error { return nil }
func (m *mockUserRepo) AssignDevices(context.Context, uint, []string, *uint) error {
	return nil
}
func (m *mockUserRepo) GetAssignedDeviceIDs(context.Context, uint) ([]string, error) {
	return nil, nil
}
func (m *mockUserRepo) IsDeviceAccessibleByUser(context.Context, uint, string) (bool, error) {
	return true, nil
}

type mockGateway struct {
	status      string
	qrCode      string
	pairingCode string
	sentMsg     string
	sentImage   []byte
	sentDoc     []byte
	presence    string
	markedRead  []string
}

func (m *mockGateway) Connect(*domainWA.WASession) error   { return nil }
func (m *mockGateway) Disconnect(uint) error               { return nil }
func (m *mockGateway) Logout(uint) error                   { return nil }
func (m *mockGateway) Purge(uint) error                    { return nil }
func (m *mockGateway) Reconnect(*domainWA.WASession) error { return nil }
func (m *mockGateway) SendMessage(_ uint, _ string, content string) error {
	m.sentMsg = content
	return nil
}
func (m *mockGateway) SendPoll(_ context.Context, _ uint, _ string, _ string, _ []string, _ int) (string, error) {
	return "poll-id", nil
}
func (m *mockGateway) SendDocument(_ context.Context, _ uint, _ string, fileBytes []byte, _, _, _ string) error {
	m.sentDoc = fileBytes
	return nil
}
func (m *mockGateway) SendImage(_ context.Context, _ uint, _ string, imageBytes []byte, _, _ string) error {
	m.sentImage = imageBytes
	return nil
}
func (m *mockGateway) SendChatPresence(_ context.Context, _ uint, _ string, state, _ string) error {
	m.presence = state
	return nil
}
func (m *mockGateway) MarkRead(_ context.Context, _ uint, ids []string, _, _ string) error {
	m.markedRead = ids
	return nil
}
func (m *mockGateway) GetStatus(uint) (string, error) {
	if m.status != "" {
		return m.status, nil
	}
	return "online", nil
}
func (m *mockGateway) GetQRCode(uint) (string, error) {
	return m.qrCode, nil
}
func (m *mockGateway) GetPairingCode(uint, string) (string, error) {
	return m.pairingCode, nil
}
func (m *mockGateway) RestoreAllSessions([]domainWA.WASession) error { return nil }

func setupTestRouter() (*http.ServeMux, *mockGateway, *mockSessionRepo) {
	repo := &mockSessionRepo{
		sessions: []domainWA.WASession{
			{ID: 1, DeviceName: "HQ Bot", PhoneNumber: "628111111111"},
		},
	}
	gw := &mockGateway{status: "online", qrCode: "qr-base64-data", pairingCode: "1234-5678"}
	resolver := ucWA.NewIdentityResolver(nil, nil, nil)
	handler := httpWA.NewHandler(repo, gw, resolver)

	mux := http.NewServeMux()
	handler.RegisterProtected(mux)
	return mux, gw, repo
}

func TestWhatsAppREST_Sessions(t *testing.T) {
	mux, _, _ := setupTestRouter()

	t.Run("GET /api/whatsapp/sessions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/sessions", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string][]httpWA.SessionDTO
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Len(t, resp["sessions"], 1)
		assert.Equal(t, "HQ Bot", resp["sessions"][0].DeviceName)
		assert.Equal(t, "online", resp["sessions"][0].Status)
	})

	t.Run("POST /api/whatsapp/sessions", func(t *testing.T) {
		body := bytes.NewBufferString(`{"device_name":"Branch Bot","phone_number":"628222222222"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/sessions", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("GET /api/whatsapp/session/status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/session/status?session_id=1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "online", resp["status"])
	})

	t.Run("GET /api/whatsapp/session/qr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/session/qr?session_id=1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "qr-base64-data", resp["qr_code"])
	})

	t.Run("POST /api/whatsapp/session/pair-code", func(t *testing.T) {
		body := bytes.NewBufferString(`{"session_id":1,"phone_number":"081234567890"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/session/pair-code", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp httpWA.PairingCodeResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "1234-5678", resp.PairingCode)
	})

	t.Run("POST /api/whatsapp/session/reconnect", func(t *testing.T) {
		body := bytes.NewBufferString(`{"session_id":1}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/session/reconnect", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST /api/whatsapp/session/logout", func(t *testing.T) {
		body := bytes.NewBufferString(`{"session_id":1}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/session/logout", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestWhatsAppREST_Messaging(t *testing.T) {
	mux, gw, _ := setupTestRouter()

	t.Run("POST /api/whatsapp/send/message", func(t *testing.T) {
		body := bytes.NewBufferString(`{"session_id":1,"to":"628123456789","content":"Halo Polyglot"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/send/message", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Halo Polyglot", gw.sentMsg)
	})

	t.Run("POST /api/whatsapp/send/image (JSON base64)", func(t *testing.T) {
		raw := []byte("fake-image-bytes")
		b64 := base64.StdEncoding.EncodeToString(raw)
		body := bytes.NewBufferString(`{"session_id":1,"to":"628123456789","image_base64":"` + b64 + `","caption":"Test"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/send/image", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, raw, gw.sentImage)
	})

	t.Run("POST /api/whatsapp/send/document (JSON base64)", func(t *testing.T) {
		raw := []byte("fake-pdf-bytes")
		b64 := base64.StdEncoding.EncodeToString(raw)
		body := bytes.NewBufferString(`{"session_id":1,"to":"628123456789","document_base64":"` + b64 + `","file_name":"invoice.pdf"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/send/document", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, raw, gw.sentDoc)
	})

	t.Run("POST /api/whatsapp/send/image (Multipart Form)", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		_ = writer.WriteField("session_id", "1")
		_ = writer.WriteField("to", "628123456789")
		part, err := writer.CreateFormFile("file", "photo.jpg")
		require.NoError(t, err)
		_, _ = part.Write([]byte("multipart-photo-bytes"))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/send/image", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []byte("multipart-photo-bytes"), gw.sentImage)
	})

	t.Run("POST /api/whatsapp/presence", func(t *testing.T) {
		body := bytes.NewBufferString(`{"session_id":1,"to":"628123456789","state":"composing","media":"text"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/presence", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "composing", gw.presence)
	})

	t.Run("POST /api/whatsapp/message/read", func(t *testing.T) {
		body := bytes.NewBufferString(`{"session_id":1,"chat_jid":"628123456789@s.whatsapp.net","message_ids":["msg-1","msg-2"]}`)
		req := httptest.NewRequest(http.MethodPost, "/api/whatsapp/message/read", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []string{"msg-1", "msg-2"}, gw.markedRead)
	})
}

func TestWhatsAppREST_Identity(t *testing.T) {
	repo := &mockSessionRepo{}
	gw := &mockGateway{}
	uRepo := &mockUserRepo{
		users: map[string]*customer.User{
			"6281299998888": {
				ID:          5,
				Username:    "agus_field",
				FullName:    "Agus Lapangan",
				Role:        "teknisi",
				PhoneNumber: "6281299998888",
				IsActive:    true,
			},
		},
	}
	resolver := ucWA.NewIdentityResolver(uRepo, nil, nil)
	handler := httpWA.NewHandler(repo, gw, resolver)

	mux := http.NewServeMux()
	handler.RegisterProtected(mux)

	t.Run("GET /api/whatsapp/identity?phone=081299998888", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/whatsapp/identity?phone=081299998888", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp httpWA.IdentityResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleTechnician, resp.Identity.Role)
		assert.Equal(t, "Agus Lapangan", resp.Identity.DisplayName)
		assert.Equal(t, uint(5), *resp.Identity.UserID)
	})
}
