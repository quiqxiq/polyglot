package provisioner

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

const planProfileComment = "AUTO plan profile"

func hotspotExpireMode(m string) port.ExpireMode {
	return port.ExpireMode(m)
}

func hotspotProfileParams(a port.SubscriberAccount) port.MikhmonProfileParams {
	return port.MikhmonProfileParams{
		Name:         a.Profile,
		AddressPool:  a.AddressPool,
		AddressList:  a.AddressList,
		SharedUsers:  intToStrOr(a.SharedUsers, "1"),
		RateLimit:    a.RateLimit,
		ParentQueue:  a.ParentQueue,
		Price:        a.Price,
		SellingPrice: a.SellingPrice,
		Validity:     a.Validity,
		ExpireMode:   hotspotExpireMode(a.ExpireMode),
		LockUser:     a.LockUser,
		LockServer:   a.LockServer,
		Comment:      planProfileComment,
	}
}

func pppProfileParams(a port.SubscriberAccount) port.PPPProfileParams {
	return port.PPPProfileParams{
		Name:          a.Profile,
		RateLimit:     a.RateLimit,
		ParentQueue:   a.ParentQueue,
		AddressList:   a.AddressList,
		LocalAddress:  a.LocalAddress,
		RemoteAddress: a.RemoteAddressPool,
		DNSServer:     a.DNSServer,
		Comment:       planProfileComment,
	}
}

func pppProfileParamsFromSpec(spec domainSub.PPPoEProfileSpec) port.PPPProfileParams {
	cmt := spec.Comment
	if cmt == "" {
		cmt = planProfileComment
	}
	return port.PPPProfileParams{
		Name:          spec.Name,
		RateLimit:     spec.RateLimit,
		LocalAddress:  spec.LocalAddress,
		RemoteAddress: spec.RemoteAddressPool,
		ParentQueue:   spec.ParentQueue,
		AddressList:   spec.AddressList,
		Comment:       cmt,
	}
}

func hotspotProfileParamsFromSpec(spec domainSub.HotspotProfileSpec) port.MikhmonProfileParams {
	cmt := spec.Comment
	if cmt == "" {
		cmt = planProfileComment
	}
	return port.MikhmonProfileParams{
		Name:           spec.Name,
		AddressPool:    spec.AddressPool,
		AddressList:    spec.AddressList,
		SharedUsers:    intToStrOr(spec.SharedUsers, "1"),
		RateLimit:      spec.RateLimit,
		ParentQueue:    spec.ParentQueue,
		SessionTimeout: spec.SessionTimeout,
		IdleTimeout:    spec.IdleTimeout,
		Comment:        cmt,
		OnLogin:        spec.OnLogin,
		OnLogout:       spec.OnLogout,
	}
}

func isolirAccount(profileName, rateLimit, addressList string) port.SubscriberAccount {
	return port.SubscriberAccount{Profile: profileName, RateLimit: rateLimit, AddressList: addressList}
}

// profileSnapshot adalah subset field profil di router yang dibandingkan
// untuk memutuskan update idempoten (F2-4).
type profileSnapshot struct {
	rate        string
	parentQueue string
	addressList string
	addressPool string
}

// planProfileDiffers melaporkan apakah profil existing berbeda dari yang
// diinginkan akun. Field yang kosong pada akun tidak dianggap perbedaan agar
// profil yang dibuat dashboard/isoli tidak tertimpa nilai kosong.
func planProfileDiffers(pr profileSnapshot, acct port.SubscriberAccount) bool {
	if acct.RateLimit != "" && pr.rate != acct.RateLimit {
		return true
	}
	if acct.ParentQueue != "" && pr.parentQueue != acct.ParentQueue {
		return true
	}
	if acct.AddressList != "" && pr.addressList != acct.AddressList {
		return true
	}
	if acct.RemoteAddressPool != "" && pr.addressPool != acct.RemoteAddressPool {
		return true
	}
	if acct.AddressPool != "" && pr.addressPool != acct.AddressPool {
		return true
	}
	return false
}

func intToStrOr(v int, def string) string {
	if v <= 0 {
		return def
	}
	return strconv.Itoa(v)
}

// EnsureProfile ensures a profile exists on the router with the given rate limit.
func (p *Provisioner) EnsureProfile(ctx context.Context, deviceID, serviceType, profileName, rateLimit string) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	_, err = p.ensurePlanProfile(ctx, driver, serviceType,
		port.SubscriberAccount{Profile: profileName, RateLimit: rateLimit})
	return err
}

// ensurePlanProfile memastikan profil akun ada di router dan mengembalikan
// nama profil AKTUAL yang dipakai router — pencocokan nama case-insensitive
// agar "isolir" vs "ISOLIR" tidak membuat profil duplikat (F1-9). Profil yang
// sudah ada di-update bila rate/address-list berbeda (F1-10/F1-11).
func (p *Provisioner) ensurePlanProfile(ctx context.Context, driver port.DeviceDriver, serviceType string, acct port.SubscriberAccount) (string, error) {
	if acct.Profile == "" || acct.RateLimit == "" {
		return acct.Profile, nil
	}
	if !isHotspot(serviceType) {
		existing, err := p.ppp.ListProfiles(ctx, driver, acct.Profile)
		if err != nil {
			return "", fmt.Errorf("list profiles: %w", err)
		}
		for _, pr := range existing {
			if strings.EqualFold(pr.Name, acct.Profile) {
				if planProfileDiffers(
					profileSnapshot{rate: pr.RateLimit, parentQueue: pr.ParentQueue, addressList: pr.AddressList, addressPool: pr.RemoteAddress},
					acct,
				) {
					params := pppProfileParams(acct)
					if _, err := p.ppp.UpdateProfile(ctx, driver, pr.RosID, params); err != nil {
						return "", fmt.Errorf("update profile %s: %w", acct.Profile, err)
					}
				}
				return pr.Name, nil
			}
		}
		if _, err := p.ppp.AddProfile(ctx, driver, pppProfileParams(acct)); err != nil {
			return "", fmt.Errorf("add profile %s: %w", acct.Profile, err)
		}
		return acct.Profile, nil
	}

	existing, err := p.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return "", fmt.Errorf("list hotspot profiles: %w", err)
	}
	for _, pr := range existing {
		if strings.EqualFold(pr.Name, acct.Profile) {
			if planProfileDiffers(
				profileSnapshot{rate: pr.RateLimit, parentQueue: pr.ParentQueue, addressList: pr.AddressList, addressPool: pr.AddressPool},
				acct,
			) {
				params := hotspotProfileParams(acct)
				if _, err := p.hot.UpdateUserProfile(ctx, driver, pr.RosID, params); err != nil {
					return "", fmt.Errorf("update hotspot profile %s: %w", acct.Profile, err)
				}
			}
			return pr.Name, nil
		}
	}
	if _, err := p.hot.CreateUserProfile(ctx, driver, hotspotProfileParams(acct)); err != nil {
		return "", fmt.Errorf("add hotspot profile %s: %w", acct.Profile, err)
	}
	return acct.Profile, nil
}

// SyncPlanProfile creates or updates the corresponding profile on the target router.
func (p *Provisioner) SyncPlanProfile(ctx context.Context, deviceID string, pl domainPlan.ServicePlan) error {
	if deviceID == "" || pl.Name == "" {
		return nil
	}
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}

	rate := pl.RateLimitWithBurst()
	if !isHotspot(pl.ServiceType) {
		remotePool := pl.RemoteAddressPool
		if pl.PPPoE != nil && pl.PPPoE.RemoteAddressPool != "" {
			remotePool = pl.PPPoE.RemoteAddressPool
		}
		addrList := pl.AddressList
		if pl.PPPoE != nil && pl.PPPoE.AddressList != "" {
			addrList = pl.PPPoE.AddressList
		}
		sessionTimeout := pl.SessionTimeout
		if pl.PPPoE != nil && pl.PPPoE.SessionTimeout != "" {
			sessionTimeout = pl.PPPoE.SessionTimeout
		}
		idleTimeout := pl.IdleTimeout
		if pl.PPPoE != nil && pl.PPPoE.IdleTimeout != "" {
			idleTimeout = pl.PPPoE.IdleTimeout
		}

		profs, err := p.ppp.ListProfiles(ctx, driver, "")
		if err != nil {
			return fmt.Errorf("list ppp profiles: %w", err)
		}
		var existing *port.PPPProfile
		for i := range profs {
			if strings.EqualFold(profs[i].Name, pl.Name) {
				existing = &profs[i]
				break
			}
		}

		params := port.PPPProfileParams{
			Name:           pl.Name,
			RateLimit:      rate,
			RemoteAddress:  remotePool,
			AddressList:    addrList,
			ParentQueue:    pl.ParentQueue,
			SessionTimeout: sessionTimeout,
			IdleTimeout:    idleTimeout,
			Comment:        "polyglot plan:" + pl.ID,
		}
		if existing != nil {
			params.LocalAddress = existing.LocalAddress
			params.OnUp = existing.OnUp
			params.OnDown = existing.OnDown
			_, err = p.ppp.UpdateProfile(ctx, driver, existing.RosID, params)
			if err != nil {
				return fmt.Errorf("update ppp profile %s: %w", pl.Name, err)
			}
		} else {
			_, err = p.ppp.AddProfile(ctx, driver, params)
			if err != nil {
				return fmt.Errorf("create ppp profile %s: %w", pl.Name, err)
			}
		}
		return nil
	}

	// Hotspot User Profile
	ipPool := pl.IPPoolName
	if pl.Hotspot != nil && pl.Hotspot.IPPoolName != "" {
		ipPool = pl.Hotspot.IPPoolName
	}
	addrList := pl.AddressList
	if pl.Hotspot != nil && pl.Hotspot.AddressList != "" {
		addrList = pl.Hotspot.AddressList
	}
	sharedUsers := 1
	if pl.Hotspot != nil && pl.Hotspot.SharedUsers > 0 {
		sharedUsers = pl.Hotspot.SharedUsers
	} else if pl.SharedUsers > 0 {
		sharedUsers = pl.SharedUsers
	}
	sessionTimeout := pl.SessionTimeout
	if pl.Hotspot != nil && pl.Hotspot.SessionTimeout != "" {
		sessionTimeout = pl.Hotspot.SessionTimeout
	}
	idleTimeout := pl.IdleTimeout
	if pl.Hotspot != nil && pl.Hotspot.IdleTimeout != "" {
		idleTimeout = pl.Hotspot.IdleTimeout
	}

	profs, err := p.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	var existing *port.HotspotUserProfile
	for i := range profs {
		if strings.EqualFold(profs[i].Name, pl.Name) {
			existing = &profs[i]
			break
		}
	}

	params := port.MikhmonProfileParams{
		Name:           pl.Name,
		RateLimit:      rate,
		AddressPool:    ipPool,
		AddressList:    addrList,
		SharedUsers:    intToStrOr(sharedUsers, "1"),
		ParentQueue:    pl.ParentQueue,
		SessionTimeout: sessionTimeout,
		IdleTimeout:    idleTimeout,
		Comment:        "polyglot plan:" + pl.ID,
	}
	if existing != nil {
		params.OnLogin = existing.OnLogin
		_, err = p.hot.UpdateUserProfile(ctx, driver, existing.RosID, params)
		if err != nil {
			return fmt.Errorf("update hotspot profile %s: %w", pl.Name, err)
		}
	} else {
		_, err = p.hot.CreateUserProfile(ctx, driver, params)
		if err != nil {
			return fmt.Errorf("create hotspot profile %s: %w", pl.Name, err)
		}
	}
	return nil
}

// DeletePlanProfile removes the corresponding profile from the router if it exists.
func (p *Provisioner) DeletePlanProfile(ctx context.Context, deviceID string, serviceType, profileName string) error {
	if deviceID == "" || profileName == "" {
		return nil
	}
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}

	if !isHotspot(serviceType) {
		profs, err := p.ppp.ListProfiles(ctx, driver, "")
		if err != nil {
			return fmt.Errorf("list ppp profiles: %w", err)
		}
		for _, pr := range profs {
			if strings.EqualFold(pr.Name, profileName) {
				_, err := p.ppp.RemoveProfile(ctx, driver, pr.RosID)
				if err != nil {
					return fmt.Errorf("remove ppp profile: %w", err)
				}
				return nil
			}
		}
		return nil
	}

	profs, err := p.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("get hotspot user profiles: %w", err)
	}
	for _, pr := range profs {
		if strings.EqualFold(pr.Name, profileName) {
			_, err := p.hot.DeleteUserProfile(ctx, driver, pr.RosID)
			if err != nil {
				return fmt.Errorf("delete hotspot user profile: %w", err)
			}
			return nil
		}
	}
	return nil
}
