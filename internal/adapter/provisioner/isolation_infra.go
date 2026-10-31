package provisioner

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domainCommand "github.com/quixiq/polyglot/internal/domain/command"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// EnsureIsolationInfrastructure configures IP pools, static DNS, profiles, NAT redirect, filter rules, and walled garden.
func (p *Provisioner) EnsureIsolationInfrastructure(ctx context.Context, deviceID string, cfg domainDevice.IsolationConfig) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	rate := cfg.RateLimit
	if rate == "" {
		rate = "0/0"
	}
	pppoeProf := cfg.PPPoEProfileName
	if pppoeProf == "" {
		pppoeProf = "ISOLIR"
	}
	hotspotProf := cfg.HotspotProfileName
	if hotspotProf == "" {
		hotspotProf = "ISOLIR"
	}
	addrList := cfg.AddressListName
	if addrList == "" {
		addrList = "ISOLIR_USERS"
	}

	localAddr := cfg.LocalAddress
	if localAddr == "" {
		localAddr = "10.100.0.1"
	}
	remotePool := cfg.RemoteAddressPool
	if remotePool == "" {
		remotePool = "pool-isolir"
	}

	// 1. Ensure dedicated IP pool for PPPoE isolation
	poolRes, err := driver.Execute(ctx, domainCommand.Command{
		Raw:  "/ip/pool/print",
		Args: map[string]string{"?name": remotePool},
	})
	if err == nil && len(poolRes.Rows) == 0 {
		_, _ = driver.Execute(ctx, domainCommand.Command{
			Raw: "/ip/pool/add",
			Args: map[string]string{
				"name":    remotePool,
				"ranges":  "10.100.0.10-10.100.0.250",
				"comment": "polyglot:isolation",
			},
		})
	}

	// 2. Ensure DNS static entries for isolation domains pointing to redirect IP
	if cfg.RedirectIP != "" {
		for _, domain := range cfg.WalledGardenDomains {
			if strings.HasSuffix(domain, ".test") || strings.HasSuffix(domain, ".isp.net") {
				cleanDom := strings.TrimPrefix(domain, "*")
				dnsRes, _ := driver.Execute(ctx, domainCommand.Command{
					Raw:  "/ip/dns/static/print",
					Args: map[string]string{"?name": cleanDom},
				})
				if len(dnsRes.Rows) == 0 {
					_, _ = driver.Execute(ctx, domainCommand.Command{
						Raw: "/ip/dns/static/add",
						Args: map[string]string{
							"name":    cleanDom,
							"address": cfg.RedirectIP,
							"comment": "polyglot:isolation",
						},
					})
				}
			}
		}
	}

	// 3. Ensure PPPoE isolation profile
	if _, err := p.ensurePlanProfile(ctx, driver, "PPPOE", port.SubscriberAccount{
		Profile:           pppoeProf,
		RateLimit:         rate,
		AddressList:       addrList,
		LocalAddress:      localAddr,
		RemoteAddressPool: remotePool,
		DNSServer:         localAddr + ",8.8.8.8",
	}); err != nil {
		return fmt.Errorf("ensure pppoe isolir profile: %w", err)
	}

	// 4. Ensure Hotspot isolation profile
	if _, err := p.ensurePlanProfile(ctx, driver, "HOTSPOT", port.SubscriberAccount{
		Profile:     hotspotProf,
		RateLimit:   rate,
		AddressList: addrList,
	}); err != nil {
		return fmt.Errorf("ensure hotspot isolir profile: %w", err)
	}

	// 5. Ensure NAT Redirect and Firewall Filter rules if configured
	if cfg.NATRedirectEnabled && cfg.RedirectIP != "" {
		redirCfg := port.IsolationRedirectConfig{
			SrcAddressList: addrList,
			PaymentHost:    cfg.RedirectIP,
			PaymentPort:    strconv.Itoa(cfg.RedirectPort),
		}
		if err := p.fw.EnsureIsolationRedirect(ctx, driver, redirCfg); err != nil {
			return fmt.Errorf("ensure isolation redirect: %w", err)
		}
		if err := p.fw.EnsureIsolationFilter(ctx, driver, addrList, cfg.RedirectIP); err != nil {
			return fmt.Errorf("ensure isolation filter: %w", err)
		}
	}

	// 6. Webproxy internal + access rules isolir (blueprint §5.4, D12).
	if err := p.ensureIsolirWebProxy(ctx, driver, cfg); err != nil {
		return fmt.Errorf("ensure isolir web proxy: %w", err)
	}
	// 7. Whitelist server billing (address-list IPBILL-GLOBAL).
	if len(cfg.BillingHosts) > 0 {
		if err := p.ensureBillingWhitelist(ctx, driver, cfg.BillingHosts); err != nil {
			return fmt.Errorf("ensure billing whitelist: %w", err)
		}
	}
	// 8. error.html mode self (best-effort, mode cloud tetap tersedia).
	p.fetchIsolirErrorPage(ctx, driver, cfg.ErrorPageURL)

	// 9. Ensure Hotspot Walled Garden for domains and portal
	if len(cfg.WalledGardenDomains) > 0 || cfg.RedirectIP != "" {
		portStr := ""
		if cfg.RedirectPort > 0 {
			portStr = strconv.Itoa(cfg.RedirectPort)
		}
		if err := p.hot.EnsureWalledGarden(ctx, driver, cfg.WalledGardenDomains, cfg.RedirectIP, portStr); err != nil {
			return fmt.Errorf("ensure walled garden: %w", err)
		}
	}
	return nil
}

// GetIsolationInfrastructureStatus returns the status of isolation profiles and rules on the router.
func (p *Provisioner) GetIsolationInfrastructureStatus(ctx context.Context, deviceID string) (domainDevice.IsolationStatus, error) {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return domainDevice.IsolationStatus{}, err
	}
	status := domainDevice.IsolationStatus{
		Config: domainDevice.DefaultIsolationConfig(),
	}

	// Check PPP profile
	pppProfiles, err := p.ppp.ListProfiles(ctx, driver, "")
	if err == nil {
		for _, prof := range pppProfiles {
			if strings.EqualFold(prof.Name, status.Config.PPPoEProfileName) {
				status.PPPoEProfileExists = true
				break
			}
		}
	}

	// Check Hotspot profile
	hotProfiles, err := p.hot.GetUserProfiles(ctx, driver)
	if err == nil {
		for _, prof := range hotProfiles {
			if strings.EqualFold(prof.Name, status.Config.HotspotProfileName) {
				status.HotspotProfileExists = true
				break
			}
		}
	}

	// Check NAT redirect rule
	hasRedirect, err := p.fw.HasIsolationRedirect(ctx, driver, status.Config.AddressListName)
	if err == nil {
		status.NATRedirectExists = hasRedirect
	}

	// Count isolated users in the address list
	count, err := p.fw.CountAddressListEntries(ctx, driver, status.Config.AddressListName)
	if err == nil {
		status.IsolatedUsersCount = count
	}

	// Blueprint §5.4: status webproxy isolir + whitelist billing.
	if res, perr := driver.Execute(ctx, domainCommand.Command{Raw: "/ip/proxy/print"}); perr == nil && len(res.Rows) > 0 {
		status.WebProxyEnabled = res.Rows[0]["enabled"] == "true"
	}
	if n, werr := p.fw.CountAddressListEntries(ctx, driver, "IPBILL-GLOBAL"); werr == nil {
		status.BillingWhitelistSize = n
	}

	return status, nil
}

// ApplyIntegrationScript updates the on-up/on-down script on the profile.
func (p *Provisioner) ApplyIntegrationScript(ctx context.Context, deviceID, profileName, serviceType, scriptType, script string) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		profs, err := p.hot.GetUserProfiles(ctx, driver)
		if err != nil {
			return fmt.Errorf("get hotspot user profiles: %w", err)
		}
		var targetProf *port.HotspotUserProfile
		for i := range profs {
			if strings.EqualFold(profs[i].Name, profileName) {
				targetProf = &profs[i]
				break
			}
		}
		if targetProf == nil {
			return fmt.Errorf("hotspot profile %q not found", profileName)
		}
		_, err = p.hot.UpdateUserProfile(ctx, driver, targetProf.RosID, port.MikhmonProfileParams{
			Name:        targetProf.Name,
			AddressPool: targetProf.AddressPool,
			SharedUsers: targetProf.SharedUsers,
			RateLimit:   targetProf.RateLimit,
			ParentQueue: targetProf.ParentQueue,
			Comment:     targetProf.Comment,
			ExpireMode:  port.ExpireModeNone,
		})
		if err != nil {
			return fmt.Errorf("update hotspot user profile: %w", err)
		}
		return nil
	}

	// PPPoE profile
	profs, err := p.ppp.ListProfiles(ctx, driver, "")
	if err != nil {
		return fmt.Errorf("list ppp profiles: %w", err)
	}
	var targetProf *port.PPPProfile
	for i := range profs {
		if strings.EqualFold(profs[i].Name, profileName) {
			targetProf = &profs[i]
			break
		}
	}
	if targetProf == nil {
		return fmt.Errorf("ppp profile %q not found", profileName)
	}
	onUp := targetProf.OnUp
	onDown := targetProf.OnDown
	switch scriptType {
	case "on-up":
		onUp = script
	case "on-down":
		onDown = script
	}
	_, err = p.ppp.UpdateProfile(ctx, driver, targetProf.RosID, port.PPPProfileParams{
		Name:          targetProf.Name,
		LocalAddress:  targetProf.LocalAddress,
		RemoteAddress: targetProf.RemoteAddress,
		RateLimit:     targetProf.RateLimit,
		ParentQueue:   targetProf.ParentQueue,
		Comment:       targetProf.Comment,
		OnUp:          onUp,
		OnDown:        onDown,
	})
	if err != nil {
		return fmt.Errorf("update ppp profile: %w", err)
	}
	return nil
}
