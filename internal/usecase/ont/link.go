package ont

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

// LinkInput menautkan ONT ke pelanggan/langganan.
type LinkInput struct {
	CustomerID     string
	SubscriptionID string
	PPPoEUsername  string
}

// Link menautkan ONT ke pelanggan dan/atau langganan (manual pin/PSB).
func (s *Service) Link(ctx context.Context, ontID string, in LinkInput) error {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return err
	}
	if in.CustomerID == "" && in.SubscriptionID == "" {
		return fmt.Errorf("%w: customer_id atau subscription_id wajib", ont.ErrInvalidInput)
	}
	if o.SubscriptionID != nil && in.SubscriptionID != "" && *o.SubscriptionID != in.SubscriptionID {
		return fmt.Errorf("%w: ONT sudah tertaut ke langganan lain", ont.ErrAlreadyLinked)
	}
	if s.deps.Customers != nil && in.CustomerID != "" {
		if _, errCust := s.deps.Customers.FindByID(ctx, in.CustomerID); errCust != nil {
			return fmt.Errorf("ont: customer %s: %w", in.CustomerID, errCust)
		}
	}
	if s.deps.Subs != nil && in.SubscriptionID != "" {
		if _, errSub := s.deps.Subs.FindByID(ctx, in.SubscriptionID); errSub != nil {
			return fmt.Errorf("ont: langganan %s: %w", in.SubscriptionID, errSub)
		}
	}
	var customerID, subscriptionID *string
	if in.CustomerID != "" {
		v := in.CustomerID
		customerID = &v
	}
	if in.SubscriptionID != "" {
		v := in.SubscriptionID
		subscriptionID = &v
	}
	if err := s.deps.ONTs.UpdateLink(ctx, ontID, customerID, subscriptionID, in.PPPoEUsername); err != nil {
		return err
	}
	s.recordEvent(ctx, ontID, "LINKED", "ONT ditautkan ke pelanggan/langganan")
	return nil
}

// Unlink melepas tautan ONT dari pelanggan/langganan.
func (s *Service) Unlink(ctx context.Context, ontID string) error {
	if _, err := s.deps.ONTs.FindByID(ctx, ontID); err != nil {
		return err
	}
	if err := s.deps.ONTs.UpdateLink(ctx, ontID, nil, nil, ""); err != nil {
		return err
	}
	s.recordEvent(ctx, ontID, "UNLINKED", "tautan ONT ke pelanggan dilepas")
	return nil
}

// PinACSDevice menyematkan device ACS ke ONT secara manual (menang atas
// heuristik serial/MAC).
func (s *Service) PinACSDevice(ctx context.Context, ontID, acsDeviceID string) error {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return err
	}
	if acsDeviceID == "" {
		return fmt.Errorf("%w: acs_device_id wajib", ont.ErrInvalidInput)
	}
	if _, errDev := s.deps.ACSRepo.FindDeviceByACSID(ctx, acsDeviceID); errDev != nil {
		return fmt.Errorf("ont: device ACS %s: %w", acsDeviceID, errDev)
	}
	if o.ACSDeviceID != nil && *o.ACSDeviceID != acsDeviceID {
		_ = s.deps.ACSRepo.UpdateDeviceMatch(ctx, *o.ACSDeviceID, nil, "")
	}
	if err := s.deps.ONTs.UpdateACS(ctx, ontID, &acsDeviceID, string(ont.MatchManual), nil, nil, 0); err != nil {
		return err
	}
	if err := s.deps.ACSRepo.UpdateDeviceMatch(ctx, acsDeviceID, &ontID, string(ont.MatchManual)); err != nil {
		return err
	}
	s.recordEvent(ctx, ontID, "ACS_PINNED", "device ACS disematkan manual: "+acsDeviceID)
	return nil
}

// UnpinACSDevice melepas pin manual device ACS dari ONT.
func (s *Service) UnpinACSDevice(ctx context.Context, ontID string) error {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return err
	}
	if o.ACSDeviceID != nil {
		_ = s.deps.ACSRepo.UpdateDeviceMatch(ctx, *o.ACSDeviceID, nil, "")
	}
	if err := s.deps.ONTs.UpdateACS(ctx, ontID, nil, "", nil, nil, 0); err != nil {
		return err
	}
	s.recordEvent(ctx, ontID, "ACS_UNPINNED", "pin manual device ACS dilepas")
	return nil
}

// ListACSDevices mengembalikan katalog device ACS (untuk UI matching).
func (s *Service) ListACSDevices(ctx context.Context, onlyUnmatched bool, search string, limit int) ([]ont.ACSDevice, error) {
	filter := port.ACSDeviceFilter{Search: search, Limit: limit}
	if onlyUnmatched {
		no := false
		filter.Matched = &no
	}
	return s.deps.ACSRepo.ListDevices(ctx, filter)
}
