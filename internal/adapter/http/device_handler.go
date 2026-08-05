package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

// DeviceHandler handles REST API requests for device inventory CRUD and live testing.
type DeviceHandler struct {
	useCase        *business.ManageDeviceUseCase
	driverProvider DriverProvider
}

// NewDeviceHandler constructs a new DeviceHandler.
func NewDeviceHandler(uc *business.ManageDeviceUseCase, provider DriverProvider) *DeviceHandler {
	return &DeviceHandler{
		useCase:        uc,
		driverProvider: provider,
	}
}

// DevicePayload defines the JSON payload for creating or updating a device record.
type DevicePayload struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Vendor         string            `json:"vendor"`
	DriverType     string            `json:"driver_type"`
	Host           string            `json:"host"`
	Port           int               `json:"port"`
	TimeoutMS      int               `json:"timeout_ms"`
	PollIntervalMS int               `json:"poll_interval_ms"`
	Extra          map[string]string `json:"extra"`
	Tags           []string          `json:"tags"`
	Enabled        bool              `json:"enabled"`
	Username       string            `json:"username"`
	Password       string            `json:"password"`
	CredExtra      map[string]string `json:"cred_extra"`
}

// CreateDevice handles POST /api/v1/devices.
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var payload DevicePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dev := device.Device{
		ID:             payload.ID,
		Name:           payload.Name,
		Vendor:         payload.Vendor,
		DriverType:     payload.DriverType,
		Host:           payload.Host,
		Port:           payload.Port,
		TimeoutMS:      payload.TimeoutMS,
		PollIntervalMS: payload.PollIntervalMS,
		Extra:          payload.Extra,
		Tags:           payload.Tags,
		Enabled:        payload.Enabled,
	}

	creds := device.Credentials{
		Username: payload.Username,
		Password: payload.Password,
		Extra:    payload.CredExtra,
	}

	if err := h.useCase.CreateDevice(c.Request.Context(), dev, creds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "device created", "device": dev})
}

// ListDevices handles GET /api/v1/devices.
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	devices, err := h.useCase.ListDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, devices)
}

// GetDevice handles GET /api/v1/devices/:deviceId.
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")
	dev, err := h.useCase.GetDevice(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dev)
}

// UpdateDevice handles PUT /api/v1/devices/:deviceId.
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")
	var payload DevicePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload.ID = deviceID

	dev := device.Device{
		ID:             payload.ID,
		Name:           payload.Name,
		Vendor:         payload.Vendor,
		DriverType:     payload.DriverType,
		Host:           payload.Host,
		Port:           payload.Port,
		TimeoutMS:      payload.TimeoutMS,
		PollIntervalMS: payload.PollIntervalMS,
		Extra:          payload.Extra,
		Tags:           payload.Tags,
		Enabled:        payload.Enabled,
	}

	creds := device.Credentials{
		Username: payload.Username,
		Password: payload.Password,
		Extra:    payload.CredExtra,
	}

	if err := h.useCase.UpdateDevice(c.Request.Context(), dev, creds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device updated", "device": dev})
}

// DeleteDevice handles DELETE /api/v1/devices/:deviceId.
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if err := h.useCase.DeleteDevice(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "device deleted"})
}

// TestConnection handles POST /api/v1/devices/:deviceId/test.
func (h *DeviceHandler) TestConnection(c *gin.Context) {
	deviceID := c.Param("deviceId")
	var driver port.DeviceDriver
	if h.driverProvider != nil {
		var err error
		driver, err = h.driverProvider(c, deviceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	result, err := h.useCase.TestConnection(c.Request.Context(), driver, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
