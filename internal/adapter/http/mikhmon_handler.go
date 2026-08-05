package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// DriverProvider function signature to obtain a port.DeviceDriver for a given deviceId.
type DriverProvider func(ctx *gin.Context, deviceID string) (port.DeviceDriver, error)

// MikhmonHandler provides HTTP REST endpoints for Mikhmon & Hotspot administration.
type MikhmonHandler struct {
	useCase               *network.MikhmonUseCase
	activeSessionsUseCase *network.ActiveSessionsUseCase
	driverProvider        DriverProvider
}

// NewMikhmonHandler constructs a new MikhmonHandler.
func NewMikhmonHandler(useCase *network.MikhmonUseCase, provider DriverProvider) *MikhmonHandler {
	return &MikhmonHandler{
		useCase:               useCase,
		activeSessionsUseCase: network.NewActiveSessionsUseCase(),
		driverProvider:        provider,
	}
}

func (h *MikhmonHandler) getDriver(c *gin.Context) (port.DeviceDriver, bool) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deviceId parameter is required"})
		return nil, false
	}
	driver, err := h.driverProvider(c, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}
	return driver, true
}

// GetDashboard returns aggregated statistics for the router dashboard.
func (h *MikhmonHandler) GetDashboard(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	summary, err := h.useCase.GetDashboardSummary(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// GetIncome returns total sales revenue calculated for today.
func (h *MikhmonHandler) GetIncome(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	income, err := h.useCase.GetTodayIncome(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"today_income": income})
}

// CreateProfile creates a new Hotspot User Profile configured for Mikhmon.
func (h *MikhmonHandler) CreateProfile(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	var params mikhmon.MikhmonProfileParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.useCase.CreateProfile(c.Request.Context(), driver, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// GetProfiles lists all Hotspot User Profiles.
func (h *MikhmonHandler) GetProfiles(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	profiles, err := h.useCase.GetProfiles(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profiles)
}

// GenerateVouchers generates a batch of hotspot vouchers.
func (h *MikhmonHandler) GenerateVouchers(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	var body struct {
		mikhmon.VoucherGenerateParams
		Count int `json:"count"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	batch, err := h.useCase.GenerateVouchers(c.Request.Context(), driver, body.VoucherGenerateParams, body.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "vouchers": batch.Vouchers})
}

// GetUsers lists hotspot users, with optional ?profile= query filter.
func (h *MikhmonHandler) GetUsers(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	profFilter := c.Query("profile")
	users, err := h.useCase.GetUsers(c.Request.Context(), driver, profFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUser fetches a single hotspot user by .id.
func (h *MikhmonHandler) GetUser(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	user, err := h.useCase.GetUser(c.Request.Context(), driver, rosID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// RemoveUser deletes a hotspot user by .id.
func (h *MikhmonHandler) RemoveUser(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	res, err := h.useCase.RemoveUser(c.Request.Context(), driver, rosID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// ResetUserCounters resets byte/time usage counters for a hotspot user.
func (h *MikhmonHandler) ResetUserCounters(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	res, err := h.useCase.ResetUserCounters(c.Request.Context(), driver, rosID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// GetActiveSessions fetches active hotspot sessions.
func (h *MikhmonHandler) GetActiveSessions(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	sessions, err := h.useCase.GetActiveSessions(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// RemoveActiveSession kicks an active hotspot session.
func (h *MikhmonHandler) RemoveActiveSession(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	res, err := h.useCase.RemoveActiveSession(c.Request.Context(), driver, rosID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// GetHosts lists hotspot hosts.
func (h *MikhmonHandler) GetHosts(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	hosts, err := h.useCase.GetHosts(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hosts)
}

// RemoveHost deletes a hotspot host entry.
func (h *MikhmonHandler) RemoveHost(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	res, err := h.useCase.RemoveHost(c.Request.Context(), driver, rosID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// GetServers lists hotspot server configurations.
func (h *MikhmonHandler) GetServers(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	servers, err := h.useCase.GetHotspotServers(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, servers)
}

// GetIPPools lists IP address pools.
func (h *MikhmonHandler) GetIPPools(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	pools, err := h.useCase.GetIPPools(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pools)
}

// GetParentQueues lists static simple queues available as parent queues.
func (h *MikhmonHandler) GetParentQueues(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	queues, err := h.useCase.GetParentQueues(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, queues)
}

// GetNATRules lists firewall NAT rules.
func (h *MikhmonHandler) GetNATRules(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rules, err := h.useCase.GetNATRules(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// SetupExpireMonitor installs or updates the Mikhmon-Expire-Monitor scheduler on the router.
func (h *MikhmonHandler) SetupExpireMonitor(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	var body struct {
		Interval string `json:"interval"`
	}
	_ = c.ShouldBindJSON(&body)
	res, err := h.useCase.SetupExpireMonitor(c.Request.Context(), driver, body.Interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// GetReports fetches sales transaction logs.
func (h *MikhmonHandler) GetReports(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	reports, err := h.useCase.GetReports(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
}

// DeleteReport deletes a sales transaction log by .id.
func (h *MikhmonHandler) DeleteReport(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	res, err := h.useCase.DeleteReport(c.Request.Context(), driver, rosID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// GetPPPActiveSessions fetches active PPPoE sessions from the router.
func (h *MikhmonHandler) GetPPPActiveSessions(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	sessions, err := h.activeSessionsUseCase.GetPPPActiveSessions(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// GetPPPInactiveSessions fetches offline PPPoE subscriber secrets from the router.
func (h *MikhmonHandler) GetPPPInactiveSessions(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	inactive, err := h.activeSessionsUseCase.GetPPPInactiveSessions(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inactive)
}

// RemovePPPActiveSession kicks an active PPPoE session by RouterOS .id.
func (h *MikhmonHandler) RemovePPPActiveSession(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	res, err := h.activeSessionsUseCase.KickPPPSession(c.Request.Context(), driver, rosID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}

// GetHotspotInactiveUsers fetches offline Hotspot users from the router.
func (h *MikhmonHandler) GetHotspotInactiveUsers(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	inactive, err := h.activeSessionsUseCase.GetHotspotInactiveUsers(c.Request.Context(), driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inactive)
}

// GetDHCPLeases fetches DHCP server leases.
func (h *MikhmonHandler) GetDHCPLeases(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	macFilter := c.Query("mac")
	leases, err := h.activeSessionsUseCase.GetDHCPLeases(c.Request.Context(), driver, macFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, leases)
}

// BlockDHCPLease blocks or unblocks a DHCP lease by RouterOS .id.
func (h *MikhmonHandler) BlockDHCPLease(c *gin.Context) {
	driver, ok := h.getDriver(c)
	if !ok {
		return
	}
	rosID := c.Param("rosId")
	var body struct {
		Blocked bool   `json:"blocked"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.activeSessionsUseCase.SetDHCPLeaseBlock(c.Request.Context(), driver, rosID, body.Blocked, body.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "result": res})
}
