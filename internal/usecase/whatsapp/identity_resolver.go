package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
func (r *IdentityResolver) Resolve(ctx context.Context, rawSender string) (*whatsapp.SenderIdentity, error) {
	norm := phone.Normalize(rawSender)
	if norm == "" {
		return nil, whatsapp.ErrInvalidPhoneNumber
	}

	jid := phone.ToWhatsAppJID(norm)

	cacheKey := fmt.Sprintf("wa:identity:%s", norm)
	if r.cache != nil {
		if cachedVal, err := r.cache.Get(ctx, cacheKey); err == nil && cachedVal != "" {
			var cachedIdent whatsapp.SenderIdentity
			if err := json.Unmarshal([]byte(cachedVal), &cachedIdent); err == nil {
				return &cachedIdent, nil
			}
		}
	}

	candidates := candidatePhoneNumbers(norm)

	// 1. Check internal users (Admin, Teknisi, etc.)
	if r.userRepo != nil {
		for _, cand := range candidates {
			u, err := r.userRepo.FindByPhone(ctx, cand)
			if err == nil && u != nil {
				var role whatsapp.SenderRole
				switch strings.ToLower(strings.TrimSpace(u.Role)) {
				case "owner", "superadmin", "admin":
					role = whatsapp.RoleAdmin
				case "teknisi", "technician":
					role = whatsapp.RoleTechnician
				default:
					role = whatsapp.RoleAdmin
				}

				name := u.FullName
				if name == "" {
					name = u.Username
				}

				identity := &whatsapp.SenderIdentity{
					PhoneNumber:   norm,
					JID:           jid,
					Role:          role,
					UserID:        &u.ID,
					DisplayName:   name,
					AllowedScopes: whatsapp.DefaultScopesForRole(role),
				}
				r.cacheIdentity(ctx, cacheKey, identity)
				return identity, nil
			}
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
	identity := &whatsapp.SenderIdentity{
		PhoneNumber:   norm,
		JID:           jid,
		Role:          whatsapp.RoleGuest,
		DisplayName:   fmt.Sprintf("Tamu (%s)", norm),
		AllowedScopes: whatsapp.DefaultScopesForRole(whatsapp.RoleGuest),
	}
	r.cacheIdentity(ctx, cacheKey, identity)
	return identity, nil
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
