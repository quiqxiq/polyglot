package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/phone"
)

const defaultIdentityCacheTTL = 300 // 5 minutes cache

// IdentityResolver resolves an inbound WhatsApp sender's phone number or JID
// to an ISP identity (Admin, Technician, Customer, or Guest) with associated
// role permissions.
type IdentityResolver struct {
	userRepo     port.UserRepository
	customerRepo port.CustomerRepository
	cache        port.CacheStore
}

// NewIdentityResolver constructs an identity resolver.
func NewIdentityResolver(
	userRepo port.UserRepository,
	customerRepo port.CustomerRepository,
	cache port.CacheStore,
) *IdentityResolver {
	return &IdentityResolver{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		cache:        cache,
	}
}

// Resolve identifies the sender from their raw phone number or WhatsApp JID.
// It may serve a cached identity and is suitable for routing and display.
func (r *IdentityResolver) Resolve(ctx context.Context, rawSender string) (*whatsapp.SenderIdentity, error) {
	return r.resolve(ctx, rawSender, false)
}

// ResolveFresh identifies the sender while bypassing cache reads. Authorization
// decisions must use this method so deactivated or re-roled staff lose access
// immediately; the fresh result still refreshes the cache.
func (r *IdentityResolver) ResolveFresh(ctx context.Context, rawSender string) (*whatsapp.SenderIdentity, error) {
	return r.resolve(ctx, rawSender, true)
}

func (r *IdentityResolver) resolve(ctx context.Context, rawSender string, fresh bool) (*whatsapp.SenderIdentity, error) {
	norm := phone.Normalize(rawSender)
	if norm == "" {
		return nil, whatsapp.ErrInvalidPhoneNumber
	}

	jid := phone.ToWhatsAppJID(norm)

	cacheKey := fmt.Sprintf("wa:identity:%s", norm)
	if !fresh && r.cache != nil {
		if cachedVal, err := r.cache.Get(ctx, cacheKey); err == nil && cachedVal != "" {
			var cachedIdent whatsapp.SenderIdentity
			if err := json.Unmarshal([]byte(cachedVal), &cachedIdent); err == nil {
				return &cachedIdent, nil
			}
		}
	}

	candidates := candidatePhoneNumbers(norm)

	// 1. Check internal users (Admin, Teknisi, CS, Finance). Only active users
	// are eligible, and a phone claimed by more than one active user is denied
	// to avoid granting privileges from ambiguous data.
	if r.userRepo != nil {
		staff, ambiguous := r.findActiveStaff(ctx, candidates)
		if ambiguous {
			logger.WithComponent("IdentityResolver").
				Warn("ambiguous whatsapp sender phone ownership")
			identity := guestIdentity(norm, jid)
			r.cacheIdentity(ctx, cacheKey, identity)
			return identity, nil
		}
		if staff != nil {
			role, known := senderRoleForUserRole(staff.Role)
			if !known {
				logger.WithComponent("IdentityResolver").
					WithField("user_id", staff.ID).
					Warn("unknown whatsapp sender role mapped to guest")
			}

			name := staff.FullName
			if name == "" {
				name = staff.Username
			}

			identity := &whatsapp.SenderIdentity{
				PhoneNumber:   norm,
				JID:           jid,
				Role:          role,
				UserID:        &staff.ID,
				DisplayName:   name,
				AllowedScopes: whatsapp.DefaultScopesForRole(role),
			}
			r.cacheIdentity(ctx, cacheKey, identity)
			return identity, nil
		}
	}

	// 2. Check registered customers (Pelanggan)
	if r.customerRepo != nil {
		for _, cand := range candidates {
			c, err := r.customerRepo.FindByPhone(ctx, cand)
			if err == nil && c.ID != "" {
				name := c.Name
				if name == "" {
					name = "Pelanggan"
				}

				identity := &whatsapp.SenderIdentity{
					PhoneNumber:   norm,
					JID:           jid,
					Role:          whatsapp.RoleCustomer,
					CustomerID:    c.ID,
					DisplayName:   name,
					AllowedScopes: whatsapp.DefaultScopesForRole(whatsapp.RoleCustomer),
				}
				r.cacheIdentity(ctx, cacheKey, identity)
				return identity, nil
			}
		}
	}

	// 3. Fallback to Guest
	identity := guestIdentity(norm, jid)
	r.cacheIdentity(ctx, cacheKey, identity)
	return identity, nil
}

// findActiveStaff returns the single active user matching any candidate phone.
// The second result reports ambiguity (more than one distinct active user).
func (r *IdentityResolver) findActiveStaff(ctx context.Context, candidates []string) (*customer.User, bool) {
	seen := make(map[uint]struct{}, len(candidates))
	matches := make([]*customer.User, 0, 1)
	for _, cand := range candidates {
		users, err := r.userRepo.FindAllByPhone(ctx, cand)
		if err != nil {
			logger.WithComponent("IdentityResolver").WithError(err).Debug("failed to look up whatsapp sender")
			continue
		}
		for _, u := range users {
			if u == nil || !u.IsActive {
				continue
			}
			if _, ok := seen[u.ID]; ok {
				continue
			}
			seen[u.ID] = struct{}{}
			matches = append(matches, u)
		}
	}
	if len(matches) > 1 {
		return nil, true
	}
	if len(matches) == 1 {
		return matches[0], false
	}
	return nil, false
}

// senderRoleForUserRole maps a stored user role to a WhatsApp sender role.
// Unknown roles deliberately fall back to guest (least privilege).
func senderRoleForUserRole(raw string) (whatsapp.SenderRole, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "owner", "superadmin", "admin":
		return whatsapp.RoleAdmin, true
	case "teknisi", "technician":
		return whatsapp.RoleTechnician, true
	case "agent", "cs":
		return whatsapp.RoleCS, true
	case "finance":
		return whatsapp.RoleFinance, true
	default:
		return whatsapp.RoleGuest, false
	}
}

func guestIdentity(norm, jid string) *whatsapp.SenderIdentity {
	return &whatsapp.SenderIdentity{
		PhoneNumber:   norm,
		JID:           jid,
		Role:          whatsapp.RoleGuest,
		DisplayName:   fmt.Sprintf("Tamu (%s)", norm),
		AllowedScopes: whatsapp.DefaultScopesForRole(whatsapp.RoleGuest),
	}
}

func (r *IdentityResolver) cacheIdentity(ctx context.Context, key string, identity *whatsapp.SenderIdentity) {
	if r.cache == nil || identity == nil {
		return
	}
	data, err := json.Marshal(identity)
	if err != nil {
		return
	}
	if err := r.cache.Set(ctx, key, string(data), defaultIdentityCacheTTL); err != nil {
		logger.WithComponent("IdentityResolver").WithError(err).Debug("failed to cache sender identity")
	}
}

// candidatePhoneNumbers derives common database variations of an Indonesian phone number.
func candidatePhoneNumbers(norm string) []string {
	candidates := []string{norm}
	if strings.HasPrefix(norm, "62") && len(norm) > 2 {
		candidates = append(candidates, "0"+norm[2:])
		candidates = append(candidates, "+"+norm)
	}
	return candidates
}
